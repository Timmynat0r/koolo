# set-pad.ps1
$env:PAD = Get-Random -Minimum 1000 -Maximum 9999
Write-Output "PAD set to $env:PAD"
"PAD=$env:PAD" | Out-File -FilePath "C:\Users\Timothy\Documents\GitHub\koolo\.vscode\pad.env" -Encoding UTF8

# Generiere eine temporäre Go-Datei mit einem zufälligen Wert
$randomHash = Get-Random -Minimum 100000 -Maximum 999999
@"
package main

var BuildRandom = $randomHash
"@ | Out-File -FilePath "C:\Users\Timothy\Documents\GitHub\koolo\cmd\koolo\random.go" -Encoding UTF8