---
title: "The most common color among permanents"
date: 2026-11-25
issues: [6207]
---
**The most common color among permanents** (#6207, CR 105.2) — "as long as red is the most common color among all permanents or is tied for most common" (the Mirage Djinns), "unless it shares a color with the most common color among all permanents or a color tied for most common" (Heroic Defiance, Tsabo's Assassin, Barrin's Unmaking) and "the most common color among nontoken permanents the chosen player controls but isn't tied" (Call to Arms). `Game.MostCommonColorsForEffect(controller, nontokenOnly)` counts each permanent once per layered colour, colourless counting for nothing, and returns every colour at the top count (none when no permanent is coloured). A layer-7c static may read it, since layer 5 has settled every permanent's colours for the pass. The card shapes are `effects.SelfPumpWhileColorIsMostCommon`, `PumpEnchantedUnlessSharesAMostCommonColor`, `SharesAMostCommonColor` and `ColorIsMostCommonUntiedAmongNontoken`. **Cards:** built by their OLD batches (b636–b846).
