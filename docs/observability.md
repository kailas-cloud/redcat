# Observability

## Деплой дашбордов

Dashboard'ы хранятся как ConfigMap с лейблом `grafana_dashboard: "1"`. Grafana автоматически подхватывает их через sidecar.

```bash
# Деплой через kustomize (вместе с основным приложением)
kubectl apply -k k8s/

# Или отдельно dashboard'ы
kubectl apply -f k8s/grafana-dashboard.yaml -f k8s/grafana-dashboard-infra.yaml
```

## Дашборды

### RedCat API (`redcat-api`)
Метрики API-сервера:
- **Total RPS** — общее количество запросов в секунду
- **P50/P99 Latency** — персентили задержки
- **Request Rate by Endpoint** — RPS по эндпоинтам
- **Latency Percentiles by Endpoint** — p50/p95/p99 по эндпоинтам
- **Requests by Status Code** — распределение по HTTP-кодам
- **POST /api/v1/places Latency** — детальная латентность добавления записей (p50/p75/p95/p99)

### RedCat Infrastructure (`redcat-infra`)
Инфраструктурные метрики нод и Valkey:

#### Node Overview
| Метрика | Источник | Зачем |
|---------|----------|-------|
| CPU Usage | `node_cpu_seconds_total` | Общая утилизация CPU ноды |
| Memory Usage | `node_memory_*` | Утилизация памяти |
| Total Memory | `node_memory_MemTotal_bytes` | Размер ноды |
| CPU Cores | `node_cpu_seconds_total` | Количество ядер |

#### CPU (Utilization & Throttling)
| Метрика | Источник | Зачем |
|---------|----------|-------|
| Node CPU Breakdown | `node_cpu_seconds_total{mode=...}` | user/system/iowait/steal — понимание куда уходит CPU |
| **CPU Throttling Ratio** | `container_cpu_cfs_throttled_*` | 🚨 Показывает, когда поды упираются в CPU limits |
| Pod CPU Usage | `container_cpu_usage_seconds_total` | Потребление CPU по подам Valkey |
| **CPU Steal Time** | `node_cpu_seconds_total{mode="steal"}` | 🔴 Noisy neighbors — гипервизор отбирает CPU для других VM |

#### Memory
| Метрика | Источник | Зачем |
|---------|----------|-------|
| Node Memory Breakdown | `node_memory_*` | used/cached/buffers/free |
| Pod Memory (working set) | `container_memory_working_set_bytes` | Реальное потребление памяти подами |

#### Disk I/O & Storage
| Метрика | Источник | Зачем |
|---------|----------|-------|
| Disk Throughput | `node_disk_{read,written}_bytes_total` | MB/s чтение/запись |
| Disk IOPS | `node_disk_{reads,writes}_completed_total` | Операции в секунду |
| PVC Usage | `kubelet_volume_stats_*` | Утилизация персистентных томов |
| PVC Used Space | `kubelet_volume_stats_used_bytes` | Абсолютные значения |

#### Network
| Метрика | Источник | Зачем |
|---------|----------|-------|
| Node Network Throughput | `node_network_{receive,transmit}_bytes_total` | Сетевой трафик ноды |
| Pod Network | `container_network_*` | Трафик по подам Valkey |
| Network Packets | `node_network_{receive,transmit}_packets_total` | PPS для анализа мелких пакетов |
| Errors & Drops | `node_network_{receive,transmit}_{drop,errs}_total` | 🔴 Проблемы сети |

## Shared Nodes: что видно

На shared нодах (Hetzner) ключевые индикаторы проблем:

1. **CPU Steal Time** (`mode="steal"`) — время, когда гипервизор отбирает CPU для других VM
   - `>5%` — заметно, стоит следить
   - `>15%` — серьёзная проблема, рассмотреть миграцию

2. **CPU Throttling** — поды упираются в limits
   - ratio `>0.1` (10%) — стоит увеличить limits или requests

3. **Memory Pressure** — OOM kills, если не влезаете в лимиты

4. **Disk I/O** — на shared storage IOPS могут быть ограничены
   - Следить за latency (если добавим `node_disk_io_time_seconds_total`)

## Дополнительные полезные метрики

Можно добавить в будущем:
- **Disk I/O Latency**: `rate(node_disk_io_time_seconds_total[5m]) / rate(node_disk_io_now[5m])`
- **Memory OOM Events**: `kube_pod_container_status_terminated_reason{reason="OOMKilled"}`
- **Pod Restarts**: `kube_pod_container_status_restarts_total`
- **Resource Requests vs Limits vs Usage** — для capacity planning
- **Valkey-specific**: `redis_*` метрики если включен exporter

## Требования

Для работы дашбордов нужны:
- **node-exporter** (метрики `node_*`)
- **kube-state-metrics** (метрики `kube_*`, `kubelet_*`)
- **cAdvisor** (метрики `container_*`) — обычно встроен в kubelet
- **Prometheus** с настроенным scrape
- **Grafana** с sidecar для автоимпорта ConfigMap с `grafana_dashboard: "1"`
