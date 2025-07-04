### Analysis

- Sending emails is I/O-bound (network) and slow per email.

- There are thousands of emails — sending sequentially would be too slow.

- Need to send emails concurrently but limit concurrency to avoid overload (throttling).

- Track and report individual send results in real-time or batch.

- Handle failures and retries per email safely.
