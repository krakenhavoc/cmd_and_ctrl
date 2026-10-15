---
title: "An enchantment that becomes an Aura as it enters"
date: 2026-10-14
issues: [3156]
---
**An enchantment that becomes an Aura as it enters** (#3156, with the flash that is sacrificed at cleanup; CR 611.2c, CR 613.1d, CR 603.4, CR 608.2b, CR 603.7, CR 514.3a, [ADR 0154](decisions/0154-auras-that-enchant-a-card-in-a-graveyard.md)'s 2026-10-14 amendment).
- **Becomes an Aura.** `effects.BecomeAnAuraAndReanimateTarget` is Necromancy's targeted enters trigger. It re-checks the source is on the battlefield (CR 603.4), adds the Aura subtype as a layer-4 effect pinned to the object (CR 611.2c), sets ADR 0154's swapped clause, returns the target under your control and attaches through `animateAttachAndSchedule`, the same steps 5 and 6 as Animate Dead. A target gone by resolution leaves the trigger doing nothing (CR 608.2b), so Necromancy stays a plain enchantment.
- **The legality check.** `attachmentLegalLocked` judges a swapped Aura attached to nothing illegal (CR 704.5m). Necromancy prints no enchant clause, so the uncatalogued-Aura exemption would otherwise have kept it.
- **Flash with a cleanup sacrifice.** `Spec.FlashWithCleanupSacrifice` (`game/flash_cleanup_sacrifice.go`) opens the instant-speed window as step 1b of the cast timing gate; a per-player restriction still closes it. `StackItem.OwnFlashCast` is stamped at announce only when a sorcery couldn't be cast and no other permission would have opened the window (ruling 2022-12-08). As the spell becomes a permanent, a `StepCleanup` delayed trigger is scheduled for that object; its current controller sacrifices it.
- **Snapshot.** `ownFlashCast` on the stack item, additive in v9.
- **Cards:** Necromancy (Full). Lightning Reflexes, Ward of Lights, Mystic Veil, Spider Climb, Parapet, Soar, Armor of Thorns, Relic Ward and Grave Servitude print the same rider and are not catalogued yet.
