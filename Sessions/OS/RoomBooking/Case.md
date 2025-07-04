| Problem                       | Risk                            |
| ----------------------------- | ------------------------------- |
| Users see same available room | Race condition                  |
| Multiple bookings at once     | Double booking (inconsistency)  |
| Room status not locked        | Concurrent writes corrupt state |
