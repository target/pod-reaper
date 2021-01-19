# pod-reaper: kills pods dead

### 2.7.0
- add `DRY_RUN` mode

### 2.6.0
- adjusted cron `SCHEDULE` for optional seconds, non optional day of week

### 2.5.0
- added multiple logging formats

### 2.4.1
- added configurable `LOG_LEVEL`

### 2.4.0
- added `UNREADY` rule to kill pods based on duration of time not passing readiness checks

### 2.3.0
- added `POD_STATUSES` rule (can now filter/kill `Evicted` pods)

### 2.2.0
- added configurable `GRACE_PERIOD` to control soft vs hard pod kills

### 2.1.0
- Added logging via [logrus](https://github.com/sirupsen/logrus)

## 2.0.0
- removed `POLL_INTERVAL` environment variable in favor of cron schedule
- added `SCHEDULE` environment variable to control when pods are inspected for reaping
  - makes use of https://godoc.org/github.com/robfig/cron
- refactored packages for clarity
- testing refactor for clarity
### 1.1.0
- added ability to only reap pods with specified labels

## 1.0.0
- redesign of the reaper to be built on modular rules
  - rules must implement two methods `load()` and `shouldReap(pod)`
  - rules determine whether or not they get loaded at runtime via environment variables
  - pods will only be reaped if all rules are met
- major refactoring for testability
