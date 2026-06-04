# log package flow

```mermaid
flowchart TD
    A["Caller uses StdLogger<br/>Trace/Debug/Info/Warn/Error/Fatal/Panic/Print"] --> B["getLogWrapper()<br/>reuse LogWrapper from sync.Pool"]
    A2["Caller uses StdLogger.WithField(s)"] --> B
    B --> C["Attach event state<br/>logger, go_id"]
    C --> C2["Run global field fillers<br/>in registration order"]
    C2 --> C3["Attach explicit fields, level,<br/>format, args"]
    C3 --> D{"Level"}

    D -->|"Trace / Debug / Info"| F["bestEffortQueue()"]
    F --> G{"logQueue has capacity?"}
    G -->|"yes"| H["enqueue wrapper"]
    G -->|"no"| I["drop low-level log<br/>putLogWrapper()<br/>dropped++"]

    D -->|"Warn / Error / Fatal / Panic"| J["mustQueue()<br/>block until enqueued"]
    J --> H

    H --> K["handleLog() goroutine<br/>range logQueue"]
    K --> L{"wrapper.done != nil?"}
    L -->|"no"| M["wrapper.flush()"]
    M --> N["format timestamp, go_id,<br/>short logger name, level, message, error"]
    N --> O["logger output<br/>default: os.Stdout"]
    O --> P["fire matching hooks<br/>with structured fields"]
    P --> R["putLogWrapper()<br/>clear fields and return to pool"]

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
    Worker->>Worker: fire hooks before barrier
    Worker-->>Log: close barrier done
    Log-->>App: return
```

## Guarantees

- `ExitTimeout` waits for logs already accepted into `logQueue` before the barrier.
- `ExitTimeout` does not close `logQueue`, so later log calls will not panic because of a closed channel.
- `Trace`, `Debug`, and `Info` are best effort and may be dropped when the queue is full.
- `Warn`, `Error`, `Fatal`, and `Panic` block until they are accepted into the queue.
- Hooks run in the log worker after console output, so accepted logs and their hooks before the barrier complete before `ExitTimeout` returns.
- Structured fields are delivered to hooks and are not rendered in console output.
- Global field fillers run in `StdLogger.wrapper()` before explicit `WithField(s)`, so explicit fields override same-key global fields.
- Multiple global field fillers are supported and run in registration order; later fillers win on same-key writes.
