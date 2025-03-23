package run

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

var baalThronePosition = data.Position{
	X: 15095,
	Y: 5042,
}

var tpPosition = data.Position{
	X: 15116,
	Y: 5071,
}

var lastClear = time.Time{}

type Baal struct {
	ctx                *context.Status
	clearMonsterFilter data.MonsterFilter // Used to clear area (basically TZ)
}

func NewBaal(clearMonsterFilter data.MonsterFilter) *Baal {
	return &Baal{
		ctx:                context.Get(),
		clearMonsterFilter: clearMonsterFilter,
	}
}

func (s Baal) Name() string {
	return string(config.BaalRun)
}

func (s Baal) Run() error {
	// Set filter
	filter := data.MonsterAnyFilter()
	if s.ctx.CharacterCfg.Game.Baal.OnlyElites {
		filter = data.MonsterEliteFilter()
	}
	if s.clearMonsterFilter != nil {
		filter = s.clearMonsterFilter
	}
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)

	err := action.WayPoint(area.TheWorldStoneKeepLevel2)
	if err != nil {
		return err
	}

	if s.ctx.CharacterCfg.Companion.Leader && s.ctx.CharacterCfg.Game.Baal.ClearFloors {
		action.OpenTPIfLeader()
		action.ClearAreaAroundPlayer(30, filter)
		action.Buff()
	}

	if s.ctx.CharacterCfg.Game.Baal.ClearFloors || s.clearMonsterFilter != nil {
		action.ClearCurrentLevel(false, filter)
	}

	err = action.MoveToArea(area.TheWorldStoneKeepLevel3)
	if err != nil {
		return err
	}

	if s.ctx.CharacterCfg.Companion.Leader && s.ctx.CharacterCfg.Game.Baal.ClearFloors {
		action.OpenTPIfLeader()
		action.ClearAreaAroundPlayer(30, filter)
		action.Buff()
	}

	if s.ctx.CharacterCfg.Game.Baal.ClearFloors || s.clearMonsterFilter != nil {
		action.ClearCurrentLevel(false, filter)
	}

	err = action.MoveToArea(area.ThroneOfDestruction)
	if err != nil {
		return err
	}

	if s.ctx.CharacterCfg.Companion.Leader && s.ctx.CharacterCfg.Game.Baal.ClearFloors {
		action.OpenTPIfLeader()
		action.ClearAreaAroundPlayer(30, filter)
		action.Buff()
		action.ClearThroughPath(tpPosition, 20, filter)
	} else {
		err = action.MoveToCoords(tpPosition)
	}

	if err != nil {
		return err
	}
	if s.checkForSoulsOrDolls() {
		return errors.New("souls or dolls detected, skipping")
	}

	// Let's move to a safe area and open the portal in companion mode
	if s.ctx.CharacterCfg.Companion.Leader {
		action.MoveToCoords(tpPosition)
		action.OpenTPIfLeader()
		action.Buff()
	}

	err = action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
	if err != nil {
		return err
	}

	// Force rebuff before waves
	action.Buff()

	// Come back to previous position
	err = action.MoveToCoords(baalThronePosition)
	if err != nil {
		return err
	}

	// Handle Baal waves
	lastClear = time.Now()
	lastWave := false

	for !lastWave {
		// Check for last wave
		if _, found := s.ctx.Data.Monsters.FindOne(npc.BaalsMinion, data.MonsterTypeMinion); found || time.Since(lastClear) > time.Minute*3 {
			lastWave = true
		}
		// Return to throne position between waves
		err = action.ClearAreaAroundPosition(baalThronePosition, 50, data.MonsterAnyFilter())
		if err != nil {
			return err
		}
		action.MoveToCoords(baalThronePosition)
		action.BuffIfRequired()
		// Small delay to allow next wave to spawn if not last wave
		if !lastWave {
			utils.Sleep(500)
		}
	}

	// Let's be sure everything is dead
	_ = action.ClearAreaAroundPosition(baalThronePosition, 50, data.MonsterAnyFilter())

	_, isLevelingChar := s.ctx.Char.(context.LevelingCharacter)
	if s.ctx.CharacterCfg.Game.Baal.KillBaal || isLevelingChar {
		utils.Sleep(15000)
		action.Buff()

		// find baal portal
		baalPortal, _ := s.ctx.Data.Objects.FindOne(object.BaalsPortal)

		// try 5 times
		maxAttempts := 5
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			err = action.InteractObject(baalPortal, func() bool {
				return s.ctx.Data.PlayerUnit.Area == area.TheWorldstoneChamber
			})

			// check if we are in worldstone chamber
			if s.ctx.Data.PlayerUnit.Area == area.TheWorldstoneChamber {

				break
			}

			// wait and try again
			utils.Sleep(1000)
			s.ctx.Logger.Debug(fmt.Sprintf("Attempt %d: Player not in TheWorldstoneChamber, retrying...", attempt))
		}

		// are we in worldstone chamber now ?
		if s.ctx.Data.PlayerUnit.Area == area.TheWorldstoneChamber {
			// kill baal
			_ = action.MoveToCoords(data.Position{X: 15136, Y: 5943})
			return s.ctx.Char.KillBaal()
		} else {
			// Log after 5 failed attempts
			s.ctx.Logger.Warn("Player still not in TheWorldstoneChamber after 5 attempts, skipping Baal kill.")
		}
	}

	return nil
}

func (s Baal) checkForSoulsOrDolls() bool {
	var npcIds []npc.ID

	if s.ctx.CharacterCfg.Game.Baal.DollQuit {
		npcIds = append(npcIds, npc.UndeadStygianDoll2, npc.UndeadSoulKiller2)
	}
	if s.ctx.CharacterCfg.Game.Baal.SoulQuit {
		npcIds = append(npcIds, npc.BlackSoul2, npc.BurningSoul2)
	}

	for _, id := range npcIds {
		if _, found := s.ctx.Data.Monsters.FindOne(id, data.MonsterTypeNone); found {
			return true
		}
	}

	return false
}
