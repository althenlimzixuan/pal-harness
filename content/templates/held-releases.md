# Held store releases

Copy to `doc/releases/held.md`. Under Company mode a phase's app store release or OTA update is held
here, and goes out as one release when all approved work is finished or when the owner asks (phase
skill, `references/ship.md`, section 8). One row per held phase; the rows are cleared once the
release is out.

| Phase | Version | Build | Needs | Notes |
|---|---|---|---|---|
| (NNN-slug) | (x.y.z) | (the store build's link, from `main`) | (`store` when the change touches native code, else `ota`) | (the phase's What to expect now lines for the phone) |
