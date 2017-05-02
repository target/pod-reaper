# pod-reaper: kills pods dead

## 2.0.0
- redesign of the reaper to be built on modular rules
    - rules must implement two methods `load()` and `shouldReap(pod)`
    - rules determine whether or not they get loaded at runtime via environment variables
    - pods will only be reaped if all rules are met
- major refactoring for testability

## 1.0.0
Initial release 