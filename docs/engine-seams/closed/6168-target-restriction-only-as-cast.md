---
title: "A target restriction that applies only as the spell is cast"
date: 2026-11-25
issues: [6168]
---
**A target restriction that applies only as the spell is cast** (#6168, CR 303.4a, 601.2c, 704.5m) — "You can't choose an untapped creature as this spell's target as you cast it" (Enthralling Hold) is `TargetSpec.TappedOnlyAsCast`, data on the enchant clause judged in `sourceAdmits` (`tappedAsCastAdmits`, game/tapped_as_cast.go) only on the walk of the card being cast (`TargetSource.casting`, set by `SourceCast`): the legal set the picker and the bot read and the announce check. The CR 608.2b re-check, a copy's new targets and the Aura's CR 704.5m enchant check (`attachmentLegalLocked`) never read it, so the spell still resolves if the creature untaps in response and the Aura stays attached after it untaps. **Cards** (1, Full): Enthralling Hold.
