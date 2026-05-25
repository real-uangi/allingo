# log package flow

```mermaid
flowchart TD
    A["Caller uses StdLogger<br/>Trace/Debug/Info/Warn/Error/Fatal/Panic/Print"] --> B["StdLogger.Basic()<br/>attach logger_name + go_id"]
    B --> C["getLogWrapper()<br/>reuse wrapper from sync.Pool"]
    C --> D["Fill wrapper<br/>level, entry, format, args"]
    D --> E{"Level"}

    E -->|"Trace / Debug / Info"| F["bestEffortQueue()"]
    F --> G{"logQueue has capacity?"}
    G -->|"yes"| H["enqueue wrapper"]
    G -->|"no"| I["drop low-level log<br/>putLogWrapper()<br/>dropped++"]

    E -->|"Warn / Error / Fatal / Panic"| J["mustQueue()<br/>block until enqueued"]
    J --> H

    H --> K["handleLog() goroutine<br/>range logQueue"]
    K --> L{"wrapper.done != nil?"}
    L -->|"no"| M["wrapper.flush()"]
    M --> N["logrus.Entry.Log / Logf"]
    N --> O["customFormatter.Format()"]
    O --> P["format timestamp, go_id,<br/>short logger name, level, message, error"]
    P --> Q["logger output<br/>default: os.Stdout"]
    Q --> R["putLogWrapper()<br/>clear fields and return to pool"]

    L -->|"yes"| S["ExitTimeout barrier reached"]
    S --> T["close(done)"]
    T --> R

    U["log.ExitTimeout(second)"] --> V["create barrier wrapper<br/>done chan struct{}"]
    V --> W{"enqueue before timeout?"}
    W -->|"yes"| H
    W -->|"no"| X["putLogWrapper()<br/>return"]
    T --> Y["ExitTimeout returns<br/>all prior queued logs flushed"]
```

## Exit sequence

```mermaid
sequenceDiagram
    participant App as app.Run
    participant Async as async pool
    participant Log as log queue
    participant Worker as handleLog goroutine
    participant Out as stdout

    App->>Async: async.ExitTimeout(5)
    Async-->>App: async tasks stopped or timed out
    App->>Log: log.ExitTimeout(5)
    Log->>Worker: enqueue barrier after existing logs
    Worker->>Out: flush queued logs before barrier
    Worker-->>Log: close barrier done
    Log-->>App: return
```

## Guarantees

- `ExitTimeout` waits for logs already accepted into `logQueue` before the barrier.
- `ExitTimeout` does not close `logQueue`, so later log calls will not panic because of a closed channel.
- `Trace`, `Debug`, and `Info` are best effort and may be dropped when the queue is full.
- `Warn`, `Error`, `Fatal`, and `Panic` block until they are accepted into the queue.
