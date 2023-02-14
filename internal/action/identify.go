package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/item"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) IdentifyAll(skipIdentify bool) *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		items := b.itemsToIdentify(data)

		if len(items) == 0 || skipIdentify {
			return
		}

		b.logger.Info("Identifying items...")
		steps = append(steps,
			step.SyncStepWithCheck(func(data game.Data) error {
				hid.PressKey(config.Config.Bindings.OpenInventory)
				return nil
			}, func(data game.Data) step.Status {
				if data.OpenMenus.Inventory {
					return step.StatusCompleted
				}
				return step.StatusInProgress
			}),
			step.SyncStep(func(data game.Data) error {
				idTome, found := getIDTome(data)
				if !found {
					b.logger.Warn("ID Tome not found, not identifying items")
					return nil
				}

				for _, i := range items {
					identifyItem(idTome, i)
				}

				hid.PressKey("esc")

				return nil
			}),
		)

		return
	}, Resettable(), CanBeSkipped())
}

func (b Builder) itemsToIdentify(data game.Data) (items []game.Item) {
	for _, i := range data.Items.Inventory {
		if i.Identified || i.Quality == item.ItemQualityNormal || i.Quality == item.ItemQualitySuperior {
			continue
		}

		items = append(items, i)
	}

	return
}

func identifyItem(idTome game.Item, i game.Item) {
	xIDTome := town.InventoryTopLeftX + idTome.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	yIDTome := town.InventoryTopLeftY + idTome.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)

	hid.MovePointer(xIDTome, yIDTome)
	helper.Sleep(200)
	hid.Click(hid.RightButton)
	helper.Sleep(200)
	x := town.InventoryTopLeftX + i.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	y := town.InventoryTopLeftY + i.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
	hid.MovePointer(x, y)
	helper.Sleep(300)
	hid.Click(hid.LeftButton)
	helper.Sleep(350)
}

func getIDTome(data game.Data) (game.Item, bool) {
	for _, i := range data.Items.Inventory {
		if i.Name == game.ItemTomeOfIdentify {
			return i, true
		}
	}

	return game.Item{}, false
}
