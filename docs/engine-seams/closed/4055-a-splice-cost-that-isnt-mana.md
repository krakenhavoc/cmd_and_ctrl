---
title: "A splice cost that isn't mana"
date: 2026-10-19
issues: [4055]
---
**A splice cost that isn't mana** (#4055, CR 702.47a, [ADR 0202](decisions/0202-splice.md) amendment 2026-10-19) closes the "A splice cost that isn't mana" row: no card waits on it now.
- **The cost.** `game.SpliceAbility.Cost` is a `SpliceCost` made of the components the abilities and additional costs already pay, one per card: a sacrifice clause, an exile-from-graveyard cost, a tap-others cost, a return-to-hand cost, or the one new `OpponentGainsLife`. The card side is `effects.SpliceOntoArcaneCost` over `SpliceSacrifice`, `SpliceExileFromGraveyard`, `SpliceTapUntapped`, `SpliceReturnToHand` and `SpliceOpponentGainsLife`; `Register` refuses an empty cost, one of both mana and something else, and a variable count.
- **The payment.** It rides `CastSpellParams.SplicePayments` (`cast_spell.splice_payments`), entry k paying `splices[k]`, and not the cast's flat lists. It is validated with the cast's other announce checks by each component's own validator, no object may be named twice by two splice cards or by one and the cast's own costs, and it is paid with the spell on the stack after the cast's own costs. A commander returned to a hand is asked about first (CR 903.9b). `SpliceOffersLocked` leaves out a card whose cost the caster cannot pay.
- **An opponent gains N life.** The caster names the opponent, from those still in the game who can gain life; the gain is `gainLifeAsCostLocked`, the CR 614 window without a pause.
- **View and client.** `splice_offers[]` carries `cost_label` and the component's picker; the cost picker shows the cost as printed and asks the opponent with the toggle, then the sacrifice, return, tap and graveyard-exile pickers ask the objects one card at a time. Bots never splice.
- **Cards** (5, all Full): Veil of Secrecy, Horobi's Whisper, Roar of Jukai, Hundred-Talon Strike and Torrent of Stone.
