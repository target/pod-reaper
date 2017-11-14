# pod-reaper: kills pods dead

## 2.0.0

- removed `POLL_INTERVAL` environment variable in favor of cron schedule
- added `SCHEDULE` environment variable to control when pods are inspected for reaping
  - makes use of https://godoc.org/github.com/robfig/cron
- refactored packages for clarity
- testing refactor for clarity

## 1.1.0

- added ability to only reap pods with specified labels

## 1.0.0

- redesign of the reaper to be built on modular rules
  - rules must implement two methods `load()` and `shouldReap(pod)`
  - rules determine whether or not they get loaded at runtime via environment variables
  - pods will only be reaped if all rules are met
- major refactoring for testability
