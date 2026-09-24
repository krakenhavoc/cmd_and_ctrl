---
title: "CR 903.9 answered from a non-battlefield zone"
date: 2026-09-17
issues: [707]
pr: 851
legacy_order: 133
---
**CR 903.9 answered from a non-battlefield zone** (#707) - the sandbox `move_card` verb (`MoveCardByIDAsCommander`) asked the question from anywhere and could only finish the move from the battlefield, so a commander moved by hand out of a graveyard, a hand, a library or the stack lost the move on both answers. Its exit half now goes through the shared exit primitive (`routeCardToZoneLocked`) like every other exit, so the one resume finishes it from whatever zone the window opened over, with the bottom-of-library position and the stack item carried across the pause on the route. Only a battlefield / stack DESTINATION is still inline - that half is an entry.
