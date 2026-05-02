# Benchmark runner para PC2 - Pipeline NLP Concurrente
# Corre 5 repeticiones por configuracion para calcular media recortada

$srcDir = Split-Path $PSScriptRoot -Parent
$repoRoot = Split-Path $srcDir -Parent

Push-Location $srcDir
try {
    $rawDir = Join-Path $repoRoot "resultados/con_results/raw"
    if (-not (Test-Path $rawDir)) {
        New-Item -ItemType Directory -Path $rawDir | Out-Null
    }

    $configs = @(
        @{ name = "n1_b1000";  token = 1;  lemma = 1;  batch = 1000 },
        @{ name = "n2_b1000";  token = 2;  lemma = 2;  batch = 1000 },
        @{ name = "n4_b1000";  token = 4;  lemma = 4;  batch = 1000 },
        @{ name = "n8_b1000";  token = 8;  lemma = 8;  batch = 1000 },
        @{ name = "n16_b1000"; token = 16; lemma = 16; batch = 1000 },
        @{ name = "n8_b100";   token = 8;  lemma = 8;  batch = 100 },
        @{ name = "n8_b5000";  token = 8;  lemma = 8;  batch = 5000 }
    )

    $repetitions = 5
    $totalRuns = $configs.Count * $repetitions
    $currentRun = 0

    Write-Host "----------------------------------------------------------------" -ForegroundColor Cyan
    Write-Host "  BENCHMARK PC2 - $($configs.Count) configs x $repetitions repeticiones = $totalRuns corridas" -ForegroundColor Cyan
    Write-Host "----------------------------------------------------------------" -ForegroundColor Cyan
    Write-Host ""

    $startTime = Get-Date

    foreach ($cfg in $configs) {
        Write-Host ">>> Configuracion: $($cfg.name) (workers=T$($cfg.token)/L$($cfg.lemma), batch=$($cfg.batch))" -ForegroundColor Yellow

        for ($i = 1; $i -le $repetitions; $i++) {
            $currentRun++
            $outFile = "$rawDir/$($cfg.name)_run$i.json"

            Write-Host "    [$currentRun/$totalRuns] Run $i/$repetitions ... " -NoNewline

            $runStart = Get-Date

            # Usar cmd /c para evitar que PowerShell interprete los logs de Go (stderr) como errores
            $goCmd = "go run ./concurrent/ -input ../data/dataset_final_1M.csv -workers-token $($cfg.token) -workers-lemma $($cfg.lemma) -batch-size $($cfg.batch) -output ../resultados/con_results/raw/$($cfg.name)_run$i.json"
            cmd /c "$goCmd >nul 2>&1"

            $runEnd = Get-Date
            $duration = ($runEnd - $runStart).TotalSeconds

            if (Test-Path $outFile) {
                $json = Get-Content $outFile -Raw | ConvertFrom-Json
                $wallStr = [math]::Round($duration, 1)
                Write-Host "$wallStr s wall, $($json.elapsed_total_ms) ms reportado" -ForegroundColor Green
            } else {
                Write-Host "FALLO - JSON no generado" -ForegroundColor Red
            }
        }
        Write-Host ""
    }

    $endTime = Get-Date
    $totalDuration = [math]::Round(($endTime - $startTime).TotalMinutes, 1)

    Write-Host "----------------------------------------------------------------" -ForegroundColor Cyan
    Write-Host "  COMPLETADO en $totalDuration minutos" -ForegroundColor Cyan
    Write-Host "  $totalRuns archivos JSON generados en $rawDir/" -ForegroundColor Cyan
    Write-Host "----------------------------------------------------------------" -ForegroundColor Cyan
}
finally {
    Pop-Location
}
