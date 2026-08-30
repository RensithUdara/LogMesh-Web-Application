param(
    [string]$ApiUrl = "http://localhost:8081",
    [int]$Requests = 1000,
    [int]$Concurrency = 20
)

$jobs = @()
for ($worker = 0; $worker -lt $Concurrency; $worker++) {
    $jobs += Start-Job -ScriptBlock {
        param($ApiUrl, $Requests, $Concurrency, $Worker)

        for ($i = $Worker; $i -lt $Requests; $i += $Concurrency) {
            $level = if ($i % 10 -eq 0) { "ERROR" } elseif ($i % 5 -eq 0) { "WARN" } else { "INFO" }
            $body = @{
                service = "load-test"
                environment = "development"
                level = $level
                message = "load test log $i"
                host = "load-worker-$Worker"
                trace_id = "load-$i"
                metadata = @{ iteration = $i }
            } | ConvertTo-Json -Depth 4

            Invoke-RestMethod -Method Post -Uri "$ApiUrl/v1/logs" -ContentType "application/json" -Body $body | Out-Null
        }
    } -ArgumentList $ApiUrl, $Requests, $Concurrency, $worker
}

$jobs | Wait-Job | Receive-Job
$jobs | Remove-Job

Invoke-RestMethod "$ApiUrl/v1/analytics"
