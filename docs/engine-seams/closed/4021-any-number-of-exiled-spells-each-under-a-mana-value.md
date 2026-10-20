---
title: "Free casts of any number of exiled spells, each under a mana value"
date: 2026-10-19
issues: [4021]
---
**Free casts of any number of exiled spells, each under a mana value** (#4021, CR 608.2g, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md) amendment 2026-10-19) — "You may cast any number of spells with mana value N or less from among them without paying their mana costs" is `effects.ExileTopWithPermission.EachManaValueAtMost`: the free, flash-speed, cast-only, pass-closed permission `TotalManaValueAtMost` builds, with `MaxSpellManaValue` as the per-spell ceiling and the new `CastPermission.AnyNumber` in place of the running budget. A cast strikes only its own card from the permission (`spendManaValueBudget`), so the rest stay castable until the holder passes; the cast path, the view's offers and the bot enumerator read the one gate (`spellManaValueWithinCap`). A land among the cards is stranded and an uncast card stays in exile. Additive in snapshot v9 (`castPermissions[].anyNumber`). **Cards:** Hazoret's Undying Fury (Full).
