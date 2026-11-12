---
title: "For Mirrodin!"
date: 2026-11-12
issues: [5654]
---
**For Mirrodin!** (#5654, CR 702.163a) — "When this Equipment enters, create a 2/2 red Rebel creature token, then attach this Equipment to it." Not an engine change: it is living weapon with a Rebel in place of the Germ. **`effects.ForMirrodin(cardName)`** (`cards/effects/for_mirrodin.go`) is the enters trigger over living weapon's shared body (`createTokenAndAttachSource`) and the existing `2/2 red Rebel` token row, read as a token the keyword's rule defines, not as words of the card's text. The attach rides the creation's keyed tail (`token-and-attach-source/attach`), so a token doubler's CR 616 order prompt is a restore point and the Equipment lands on the first Rebel made; an Equipment that left in response, or came back as a new object (CR 400.7), is attached to nothing and the 2/2 stays. No new effect key, token row, snapshot field or wire field.

**Cards:** Barbed Batterfist, Dragonwing Glider, Goldwarden's Helm, Hexgold Halberd, Hexgold Hoverwings, Kemba's Banner, Mirran Bardiche, Sylvok Battle-Chair, Vulshok Splitter (all Full).
