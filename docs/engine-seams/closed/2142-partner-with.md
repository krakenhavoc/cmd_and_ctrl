---
title: "Partner with"
date: 2026-10-09
issues: [2142]
---
**Partner with** (#2142, CR 702.124j) — `deck/partner.go` is the deck-construction half, `effects/partner_with.go` the entry trigger.
- **Two commanders.** `deck.Validate` accepts two commanders when each prints a "Partner with [name]" line naming the other (read off Scryfall's oracle text, so it covers every card, catalog or not). Only one naming the other, or each naming a third card, is `invalid_partner_pair`, naming the card at fault; two commanders with no partner-with between them stay `too_many_commanders`, and so does a third commander (CR 702.124g). Each commander is checked as a legal commander on its own, and the deck's colour identity is the pair's union (CR 702.124c). A partner-with card alone is an ordinary commander. The game already kept commander tax and commander damage per commander (CR 702.124d), so nothing changed there.
- **The entry search.** `effects.PartnerWith(card, partner)` is "When this permanent enters, target player may search their library for a card named [name], reveal it, put it into their hand, then shuffle": a `TargetPlayer` trigger (any player), a `MayChoice` addressed to the target, so declining searches nothing and shuffles nothing, then an `Optional` `SearchLibrary` of their own library by `game.HasName`, so they may still fail to find (CR 701.23b). The row declares its Effect, so a trigger waiting on the stack is a restore point.
- **Not built:** plain partner, partner—[text] (Friends forever and the rest), choose a Background and Doctor's companion are still refused at deck import, as before.
- **Cards (8):** Frodo, Adventurous Hobbit, Sam, Loyal Attendant, Toothy, Imaginary Friend, Ley Weaver, Lore Weaver, Blaring Captain and Blaring Recruiter, Full; Pir, Imaginative Rascal's caveat came off.
