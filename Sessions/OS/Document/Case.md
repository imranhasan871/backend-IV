### Analysis

- Conflict will be less
- Concurrent read can be allowed
- Write must be done by single thread


Scenario (no versioning, but mutex used):

    Alice reads the doc → "Version 3", content: "Initial"

    Bob also reads → Version 3, content: "Initial"

    Alice saves: "Alice's content" → now Version 4

    Bob saves: "Bob's content"

➡️ Result: Bob’s write overwrites Alice’s.
No error, even though Alice wrote after Bob read.

🟥 This is a classic lost update — no data race, but a logic bug.
🟥 Mutex alone can’t detect that Bob is saving stale data.