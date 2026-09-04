param(
  [Parameter(Mandatory)][string]$Label,      # e.g. "rq1-replicas8-run1"
  [int]$WindowSeconds = 180,
  [long]$EndEpoch = 0,                       # 0 = now
  [string]$Prometheus = "http://localhost:9000",
  [string]$OutFile = "results.csv"
)

if ($EndEpoch -eq 0) { $EndEpoch = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds() }
$W = "${WindowSeconds}s"
$S = $WindowSeconds

# metric name -> PromQL. $W is the window, $S the window in seconds.
$queries = [ordered]@{
  # ---- HTTP, per service
  "throughput_rps"        = "sum by (job) (increase(http_requests_total[$W])) / $S"
  "errors_rps"            = "sum by (job) (increase(http_requests_total{status=~`"5..`"}[$W])) / $S"
  "error_ratio"           = "sum by (job) (increase(http_requests_total{status=~`"5..`"}[$W])) / sum by (job) (increase(http_requests_total[$W]))"
  "latency_p50_s"         = "histogram_quantile(0.50, sum by (job, le) (rate(http_request_duration_seconds_bucket[$W])))"
  "latency_p95_s"         = "histogram_quantile(0.95, sum by (job, le) (rate(http_request_duration_seconds_bucket[$W])))"
  "latency_p99_s"         = "histogram_quantile(0.99, sum by (job, le) (rate(http_request_duration_seconds_bucket[$W])))"
  "in_flight_avg"         = "avg_over_time(http_requests_in_flight[$W])"
  "in_flight_max"         = "max_over_time(http_requests_in_flight[$W])"

  # ---- database queries
  "db_query_rps"          = "sum by (job) (increase(db_queries_total[$W])) / $S"
  "db_query_errors"       = "sum by (job) (increase(db_queries_total{outcome=`"error`"}[$W]))"
  "db_query_p95_s"        = "histogram_quantile(0.95, sum by (job, le) (rate(db_query_duration_seconds_bucket[$W])))"

  # ---- connection pool  (the RQ1 evidence)
  "pool_max"              = "max_over_time(db_pool_max_conns[$W])"
  "pool_acquired_avg"     = "avg_over_time(db_pool_acquired_conns[$W])"
  "pool_acquired_max"     = "max_over_time(db_pool_acquired_conns[$W])"
  "pool_utilisation_avg"  = "avg_over_time(db_pool_acquired_conns[$W]) / max_over_time(db_pool_max_conns[$W])"
  "pool_empty_acquires"   = "increase(db_pool_empty_acquire_total[$W])"
  "pool_wait_total_s"     = "increase(db_pool_empty_acquire_wait_seconds_total[$W])"
  "pool_wait_mean_s"      = "increase(db_pool_empty_acquire_wait_seconds_total[$W]) / increase(db_pool_empty_acquire_total[$W])"

  # ---- resources, per service
  "cpu_cores_avg"         = "increase(process_cpu_seconds_total[$W]) / $S"
  "memory_mb_avg"         = "avg_over_time(process_resident_memory_bytes[$W]) / 1024 / 1024"
  "memory_mb_max"         = "max_over_time(process_resident_memory_bytes[$W]) / 1024 / 1024"
  "goroutines_max"        = "max_over_time(go_goroutines[$W])"

  # ---- async path
  "clicks_published"      = "increase(clicks_published_total[$W])"
  "clicks_dropped"        = "increase(clicks_dropped_total[$W])"
  "clicks_consumed"       = "increase(clicks_consumed_total{outcome=`"ok`"}[$W])"
  "clicks_buffer_util"    = "avg_over_time(clicks_buffer_used[$W]) / max_over_time(clicks_buffer_capacity[$W])"
  "queue_depth_avg"       = "avg_over_time(rabbitmq_queue_messages_ready[$W])"
  "queue_depth_max"       = "max_over_time(rabbitmq_queue_messages_ready[$W])"
  "flush_p95_s"           = "histogram_quantile(0.95, sum by (le) (rate(clicks_flush_duration_seconds_bucket[$W])))"

  # ---- postgres, server side
  "pg_backends_avg"       = "avg_over_time(pg_stat_database_numbackends{datname=`"urlshortener`"}[$W])"
  "pg_backends_max"       = "max_over_time(pg_stat_database_numbackends{datname=`"urlshortener`"}[$W])"
  "pg_commits_per_s"      = "increase(pg_stat_database_xact_commit{datname=`"urlshortener`"}[$W]) / $S"
  "pg_cache_hit_ratio"    = "sum(increase(pg_stat_database_blks_hit[$W])) / (sum(increase(pg_stat_database_blks_hit[$W])) + sum(increase(pg_stat_database_blks_read[$W])))"

  # ---- replica count, recorded automatically
  "replicas"              = "count by (job) (up == 1)"
}

$rows = foreach ($name in $queries.Keys) {
  $uri = "$Prometheus/api/v1/query?query=$([uri]::EscapeDataString($queries[$name]))&time=$EndEpoch"
  try {
    $resp = Invoke-RestMethod -Uri $uri -TimeoutSec 20
  } catch {
    Write-Warning "$name : request failed"
    continue
  }
  if ($resp.status -ne 'success' -or $resp.data.result.Count -eq 0) {
    Write-Warning "$name : no data"
    continue
  }
  foreach ($series in $resp.data.result) {
    [pscustomobject]@{
      run        = $Label
      end_epoch  = $EndEpoch
      window_s   = $WindowSeconds
      metric     = $name
      job        = $(if ($series.metric.job) { $series.metric.job } else { "-" })
      value      = [double]$series.value[1]
    }
  }
}

$exists = Test-Path $OutFile
$rows | Export-Csv -Path $OutFile -NoTypeInformation -Append:$exists
Write-Host "$Label : wrote $($rows.Count) rows to $OutFile"
