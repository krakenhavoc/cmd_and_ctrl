---
title: "Tap N untapped permanents you control as an ability cost"
date: 2026-09-24
issues: [758]
pr: 1426
---
**Tap N untapped permanents you control as an ability cost** (#758, CR 118.3 / 302.6 / 602.2b) — one fixed-count `game.TapOthersCost` with a count, non-targeting `TargetSpec` filter, printed-word `ExcludeSource` and prompt label, shared by `AbilityCost` and `ManaAbilityShape`. Both activation paths validate every named permanent before paying anything, refuse duplicates, tapped or uncontrolled permanents and a source already spent by `{T}`, then emit one tap event per payment. Both wire views expose `tap_others_label` / `tap_others_options`; both actions accept `tap_ids`; the legal enumerator expands the same candidate walk and the heuristic prices the lost attack/block; the client reuses the sacrifice modal with the verb "Tap". The auto-tapper excludes named payers and never plans a mana source whose own activation needs this choice. **Cards:** the station family and The Shire already exercised the ordinary activated path; Scene of the Crime, Survivors' Encampment, Springleaf Drum, Jaspera Sentinel, Heritage Druid, Relic of Legends and Holdout Settlement exercise the mana path and are `full`. The variable-count follow-up is closed separately above; fixed-count catalog adoption continues in the owning batch issues rather than reopening this engine seam.
