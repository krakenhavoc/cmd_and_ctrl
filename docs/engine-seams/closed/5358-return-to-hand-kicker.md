---
title: "A kicker that returns a permanent you control"
date: 2026-11-09
issues: [5358]
---
**A kicker that returns a permanent you control** (#5358, CR 702.33a / 601.2b / 601.2g / 601.2h / 118.3 / 707.10) — #5247 put `game.AdditionalCost.ReturnToHand` on a cast's mandatory slot, but `effects.Register` refused it on an optional one, so Chocobo Kick ("Kicker—Return a land you control to its owner's hand") could not be built. `effects.KickerReturnToHand(label, preds...)` is that same component on the optional slot, as `KickerSacrifice` is for a sacrifice, so the candidate walk, payability, validator and payer are #5247's.
- **The cast.** The component joins the plan only when the kicker is announced (CR 601.2b); from there `planReturnToHand` reads it from either slot, and it is validated, asked CR 903.9b about and paid with the spell on the stack exactly as the mandatory form. Unkicked, a `return_ids` is refused. A land tapped for the spell's own mana may still pay it (CR 601.2g before 601.2h).
- **The wire.** Additive: `optional_costs[i].return_label` / `return_options`, the mandatory slot's picker view off the same walk; fewer cards than `min` means the offer cannot be taken, and the unkicked cast stays legal. The client greys the offer when nothing can pay and asks for the land in the existing return step (`castReturnCost` reads a claimed kicker's picker).
- **The bot.** `returnCastPayment` takes the announced optional set, so the enumerator offers a kicked cast naming one land on `return_ids` and drops only the kicked announcement when there is nothing to return.
- **Guards.** `effects.Register` holds the optional component to the activated form's clause check and refuses a second return-to-hand component on one card (mandatory beside optional, or two optional): a cast names them all on one `return_ids` list. A card-shaped kicker is paid at most once, as before.
- **Cards:** Chocobo Kick (Full).
