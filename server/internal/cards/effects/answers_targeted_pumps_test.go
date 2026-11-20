package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answers_targeted_pumps_test.go — ADR 0009's amendment of 2026-10-30
// (#4734), slice 7's second batch (ADR 0142's amendment of 2026-11-05):
// the targeted protective, pump and prevent rows declare what they give
// their target, so smart autopass matches each move by its kind
// (answers_now) instead of stopping at every opponent item and key
// window through has_targets. Each row declares the union of what it
// gives (ADR 0142 sweep ruling 4): regenerate and indestructible are
// survive, shroud, hexproof and protection are protect, +N/+N and
// +1/+1 counters are pump, a damage shield or a fog on one creature is
// prevent, and a combat keyword or "can't be blocked" beside a pump is
// combat_grant. A cost that sacrifices another creature adds sac_outlet
// (the guard's cost coupling); sacrificing the row's own source does not,
// since owner answer 4 already reads it from the game.
//
// Left undeclared, on purpose: a row that can be pointed at an
// opponent's creature as removal (+1/-1, -1/-1, +X/-X), whose `remove`
// would carry no verdict anyway, and a row that also makes a permanent a
// creature for good (Case of the Filched Falcon, Titania).

// TestTargetedPumpAndPreventRowsDeclare pins every row of the batch, by
// card and printed label.
func TestTargetedPumpAndPreventRowsDeclare(t *testing.T) {
	for _, r := range []struct {
		card, label string
		want        game.Answers
	}{
		{"Accursed Duneyard", "{2}, {T}: Regenerate target Shade, Skeleton, Specter, Spirit, Vampire, Wraith, or Zombie.", game.AnswerSurvive},
		{"Agatha's Soul Cauldron", "{T}: Exile target card from a graveyard. When a creature card is exiled this way, put a +1/+1 counter on target creature you control.", game.AnswerPump | game.AnswerGraveyard},
		{"Alaborn Veteran", "{T}: Target creature gets +2/+2 until end of turn. Activate only during your turn, before attackers are declared.", game.AnswerPump},
		{"Amok", "{1}, Discard a card at random: Put a +1/+1 counter on target creature.", game.AnswerPump},
		{"Angelic Page", "{T}: Target attacking or blocking creature gets +1/+1 until end of turn.", game.AnswerPump},
		{"Animal Sanctuary", "{2}, {T}: Put a +1/+1 counter on target Bird, Cat, Dog, Goat, Ox, or Snake.", game.AnswerPump},
		{"Asceticism", "{1}{G}: Regenerate target creature.", game.AnswerSurvive},
		{"Ashnod's Battle Gear", "{2}, {T}: Target creature you control gets +2/-2 for as long as this artifact remains tapped.", game.AnswerPump},
		{"Avacyn, Guardian Angel", "{1}{W}: Prevent all damage that would be dealt to another target creature this turn by sources of the color of your choice.", game.AnswerPrevent},
		{"Avacyn, Guardian Angel", "{5}{W}{W}: Prevent all damage that would be dealt to target player or planeswalker this turn by sources of the color of your choice.", game.AnswerPrevent},
		{"Avatar Roku, Firebender", "{R}{R}{R}: Target creature gets +3/+0 until end of turn.", game.AnswerPump},
		{"Baton of Courage", "Remove a charge counter from this artifact: Target creature gets +1/+1 until end of turn.", game.AnswerPump},
		{"Bearer of Memory", "{5}{G}: Put a +1/+1 counter on target enchantment creature. It gains trample until end of turn.", game.AnswerPump | game.AnswerCombatGrant},
		{"Benalish Missionary", "{1}{W}, {T}: Prevent all combat damage that would be dealt by target blocked creature this turn.", game.AnswerPrevent},
		{"Benevolent Hydra", "{T}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on another target creature you control.", game.AnswerPump},
		{"Blinkmoth Nexus", "{1}, {T}: Target Blinkmoth creature gets +1/+1 until end of turn.", game.AnswerPump},
		{"Boneknitter", "{1}{B}: Regenerate target Zombie.", game.AnswerSurvive},
		{"Broken Fall", "Return this enchantment to its owner's hand: Regenerate target creature.", game.AnswerSurvive},
		{"Captive Flame", "{R}: Target creature gets +1/+0 until end of turn.", game.AnswerPump},
		{"Cenn's Tactician", "{W}, {T}: Put a +1/+1 counter on target Soldier creature.", game.AnswerPump},
		{"Centaur Garden", "Threshold — {G}, {T}, Sacrifice this land: Target creature gets +3/+3 until end of turn. Activate only if there are seven or more cards in your graveyard.", game.AnswerPump},
		{"Cephalid Illusionist", "{2}{U}, {T}: Prevent all combat damage that would be dealt to and dealt by target creature you control this turn.", game.AnswerPrevent},
		{"Collector's Cage", "{1}, {T}: Put a +1/+1 counter on target creature you control. Then if you control three or more creatures with different powers, you may play the exiled card without paying its mana cost.", game.AnswerPump},
		{"Coral Helm", "{3}, Discard a card at random: Target creature gets +2/+2 until end of turn.", game.AnswerPump},
		{"Crowd of True Believers", "{T}: Target creature you control that's attacking alone gets +1/+0 until end of turn. You gain 1 life.", game.AnswerPump},
		{"Daring Mechanic", "{3}{W}: Put a +1/+1 counter on target Mount or Vehicle.", game.AnswerPump},
		{"Dawnhart Mentor", "Coven — {5}{G}: Target creature you control gets +3/+3 and gains trample until end of turn. Activate only if you control three or more creatures with different powers.", game.AnswerPump | game.AnswerCombatGrant},
		{"Defender en-Vec", "Remove a fade counter from this creature: Prevent the next 2 damage that would be dealt to any target this turn.", game.AnswerPrevent},
		{"Delif's Cube", "{2}, Remove a cube counter from this artifact: Regenerate target creature.", game.AnswerSurvive},
		{"Dina, Essence Brewer", "{2}, {T}, Sacrifice another creature: You gain X life and put X +1/+1 counters on target creature you control, where X is the sacrificed creature's power", game.AnswerPump | game.AnswerSacOutlet},
		{"Dr. Madison Li", "{T}, Pay {E}: Target creature gets +1/+0 and gains trample and haste until end of turn.", game.AnswerPump | game.AnswerCombatGrant},
		{"Draconian Cylix", "{2}, {T}, Discard a card at random: Regenerate target creature.", game.AnswerSurvive},
		{"Dragon Blood", "{3}, {T}: Put a +1/+1 counter on target creature.", game.AnswerPump},
		{"Drannith Ruins", "{2}, {T}: Put two +1/+1 counters on target non-Human creature that entered this turn.", game.AnswerPump},
		{"Dwarven Lieutenant", "{1}{R}: Target Dwarf creature gets +1/+0 until end of turn.", game.AnswerPump},
		{"Earth Kingdom Protectors", "Sacrifice this creature: Another target Ally you control gains indestructible until end of turn.", game.AnswerSurvive},
		{"Ebony Horse", "{2}, {T}: Untap target attacking creature you control. Prevent all combat damage that would be dealt to and dealt by that creature this turn.", game.AnswerPrevent},
		{"Elven Lyre", "{1}, {T}, Sacrifice this artifact: Target creature gets +2/+2 until end of turn.", game.AnswerPump},
		{"Elvish Scout", "{G}, {T}: Untap target attacking creature you control. Prevent all combat damage that would be dealt to and dealt by it this turn.", game.AnswerPrevent},
		{"Endoskeleton", "{2}, {T}: Target creature gets +0/+3 for as long as this artifact remains tapped.", game.AnswerPump},
		{"Everglove Courier", "{2}{G}, {T}: Target Elf creature gets +2/+2 and has trample for as long as this creature remains tapped.", game.AnswerPump | game.AnswerCombatGrant},
		{"Experiment Kraj", "{T}: Put a +1/+1 counter on target creature.", game.AnswerPump},
		{"Extruder", "Sacrifice an artifact: Put a +1/+1 counter on target creature.", game.AnswerPump | game.AnswerSacOutlet},
		{"Fain, the Broker", "{T}, Sacrifice a creature: Put two +1/+1 counters on target creature.", game.AnswerPump | game.AnswerSacOutlet},
		{"Flamestick Courier", "{2}{R}, {T}: Target Goblin creature gets +2/+2 and has haste for as long as this creature remains tapped.", game.AnswerPump | game.AnswerCombatGrant},
		{"Frightshroud Courier", "{2}{B}, {T}: Target Zombie creature gets +2/+2 and has fear for as long as this creature remains tapped. (It can't be blocked except by artifact creatures and/or black creatures.)", game.AnswerPump | game.AnswerCombatGrant},
		{"Gallia, the Merrymaker", "{1}{R}, {T}: Put a +1/+1 counter on target creature that entered this turn.", game.AnswerPump},
		{"Ghosthelm Courier", "{2}{U}, {T}: Target Wizard creature gets +2/+2 and has shroud for as long as this creature remains tapped. (It can't be the target of spells or abilities.)", game.AnswerPump | game.AnswerProtect},
		{"Goblin Chirurgeon", "Sacrifice a Goblin: Regenerate target creature.", game.AnswerSurvive | game.AnswerSacOutlet},
		{"Godtoucher", "{1}{W}, {T}: Prevent all damage that would be dealt to target creature with power 5 or greater this turn.", game.AnswerPrevent},
		{"Golgari Guildmage", "{4}{G}: Put a +1/+1 counter on target creature.", game.AnswerPump},
		{"Gossamer Chains", "Return this enchantment to its owner's hand: Prevent all combat damage that would be dealt by target unblocked creature this turn.", game.AnswerPrevent},
		{"Guard Dogs", "{2}{W}, {T}: Choose a permanent you control. Prevent all combat damage target creature would deal this turn if it shares a color with that permanent.", game.AnswerPrevent},
		{"Gyome, Master Chef", "{1}, Sacrifice a Food: Target creature gains indestructible until end of turn. Tap it.", game.AnswerSurvive},
		{"Heart of Yavimaya", "{T}: Target creature gets +1/+1 until end of turn.", game.AnswerPump},
		{"Herald of the Sun", "{3}{W}: Put a +1/+1 counter on another target creature with flying.", game.AnswerPump},
		{"Hidden Retreat", "Put a card from your hand on top of your library: Prevent all damage that would be dealt by target instant or sorcery spell this turn.", game.AnswerPrevent},
		{"High Sentinels of Arashin", "{3}{W}: Put a +1/+1 counter on target creature.", game.AnswerPump},
		{"Hisoka's Guard", "{1}{U}, {T}: Target creature you control other than this creature has shroud for as long as this creature remains tapped. (It can't be the target of spells or abilities.)", game.AnswerProtect},
		{"Horn of Deafening", "{2}, {T}: Prevent all combat damage that would be dealt by target creature this turn.", game.AnswerPrevent},
		{"Hound Tamer", "{3}{G}: Put a +1/+1 counter on target creature.", game.AnswerPump},
		{"Icatian Lieutenant", "{1}{W}: Target Soldier creature gets +1/+0 until end of turn.", game.AnswerPump},
		{"Icatian Priest", "{1}{W}{W}: Target creature gets +1/+1 until end of turn.", game.AnswerPump},
		{"Ichiga, Who Topples Oaks", "Remove a ki counter from Ichiga: Target creature gets +2/+2 until end of turn.", game.AnswerPump},
		{"Immaculate Magistrate", "{T}: Put a +1/+1 counter on target creature for each Elf you control.", game.AnswerPump},
		{"Infantry Veteran", "{T}: Target attacking creature gets +1/+1 until end of turn.", game.AnswerPump},
		{"Ingenious Leonin", "{3}{W}: Put a +1/+1 counter on another target attacking creature you control. If that creature is a Cat, it gains first strike until end of turn.", game.AnswerPump | game.AnswerCombatGrant},
		{"Ith, High Arcanist", "{T}: Untap target attacking creature. Prevent all combat damage that would be dealt to and dealt by that creature this turn.", game.AnswerPrevent},
		{"Jadzia Dax, Living Legacy", "Transfer Symbiont — {2}{G}, Discard this card: Put two +1/+1 counters on target creature.", game.AnswerPump},
		{"Kaiso, Memory of Loyalty", "Remove a ki counter from Kaiso: Prevent all damage that would be dealt to target creature this turn.", game.AnswerPrevent},
		{"Kessig Wolf Run", "{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn", game.AnswerPump | game.AnswerCombatGrant},
		{"Kitsune Healer", "{T}: Prevent the next 1 damage that would be dealt to any target this turn.", game.AnswerPrevent},
		{"Kitsune Healer", "{T}: Prevent all damage that would be dealt to target legendary creature this turn.", game.AnswerPrevent},
		{"Kor Haven", "{1}{W}, {T}: Prevent all combat damage that would be dealt by target attacking creature this turn.", game.AnswerPrevent},
		{"Kry Shield", "{2}, {T}: Prevent all damage that would be dealt this turn by target creature you control. That creature gets +0/+X until end of turn, where X is its mana value.", game.AnswerPrevent | game.AnswerPump},
		{"Kurbis, Harvest Celebrant", "Remove a +1/+1 counter from Kurbis: Prevent all damage that would be dealt this turn to another target creature with a +1/+1 counter on it.", game.AnswerPrevent},
		{"Lady Evangela", "{W}{B}, {T}: Prevent all combat damage that would be dealt by target creature this turn.", game.AnswerPrevent},
		{"Martyr of Spores", "{1}, Reveal X green cards from your hand, Sacrifice this creature: Target creature gets +X/+X until end of turn.", game.AnswerPump},
		{"Maze of Ith", "{T}: Untap target attacking creature. Prevent all combat damage that would be dealt to and dealt by that creature this turn.", game.AnswerPrevent},
		{"Maze of Shadows", "{T}: Untap target attacking creature with shadow. Prevent all combat damage that would be dealt to and dealt by that creature this turn.", game.AnswerPrevent},
		{"Mishra's Factory", "{T}: Target Assembly-Worker creature gets +1/+1 until end of turn.", game.AnswerPump},
		{"Mishra's Foundry", "{1}, {T}: Target attacking Assembly-Worker gets +2/+2 until end of turn.", game.AnswerPump},
		{"Molting Skin", "Return this enchantment to its owner's hand: Regenerate target creature.", game.AnswerSurvive},
		{"Okina, Temple to the Grandfathers", "{G}, {T}: Target legendary creature gets +1/+1 until end of turn.", game.AnswerPump},
		{"Old Hob, Alleycat Blues", "{1}{W}: Target attacking creature token gains indestructible until end of turn.", game.AnswerSurvive},
		{"Ollenbock Escort", "Sacrifice this creature: Target creature you control with a +1/+1 counter on it gains lifelink and indestructible until end of turn.", game.AnswerSurvive | game.AnswerCombatGrant},
		{"Orcish Healer", "{B}{B}{R}, {T}: Regenerate target black or green creature.", game.AnswerSurvive},
		{"Orcish Healer", "{R}{G}{G}, {T}: Regenerate target black or green creature.", game.AnswerSurvive},
		{"Oriss, Samite Guardian", "{T}: Prevent all damage that would be dealt to target creature this turn.", game.AnswerPrevent},
		{"Pang Tong, \"Young Phoenix\"", "{T}: Target creature gets +0/+2 until end of turn. Activate only during your turn, before attackers are declared.", game.AnswerPump},
		{"Patriot, Shield Wielder", "{2}, {T}: Another target creature you control gets +2/+0 and gains hexproof until end of turn.", game.AnswerPump | game.AnswerProtect},
		{"Pearlspear Courier", "{2}{W}, {T}: Target Soldier creature gets +2/+2 and has vigilance for as long as this creature remains tapped.", game.AnswerPump | game.AnswerCombatGrant},
		{"Pia, Determined Rebuilder", "{5}{R}: Target creature gets +X/+0 until end of turn, where X is the number of artifacts you control", game.AnswerPump},
		{"Proficient Pyrodancer", "{2}{R}: Another target creature with an art sticker on it gets +2/+0 and gains menace until end of turn.", game.AnswerPump | game.AnswerCombatGrant},
		{"Pyrewild Shaman", "Bloodrush — {1}{R}, Discard this card: Target attacking creature gets +3/+1 until end of turn.", game.AnswerPump},
		{"Resistance Fighter", "Sacrifice this creature: Prevent all combat damage target creature would deal this turn.", game.AnswerPrevent},
		{"Rhonas the Indomitable", "{2}{G}: Another target creature gets +2/+0 and gains trample until end of turn.", game.AnswerPump | game.AnswerCombatGrant},
		{"Safeguard", "{2}{W}: Prevent all combat damage that would be dealt by target creature this turn.", game.AnswerPrevent},
		{"Samite Sanctuary", "{2}: Prevent the next 1 damage that would be dealt to target creature this turn. Any player may activate this ability.", game.AnswerPrevent},
		{"Selfless Samurai", "Sacrifice this creature: Another target creature you control gains indestructible until end of turn.", game.AnswerSurvive},
		{"Selfless Savior", "Sacrifice Selfless Savior: Another target creature you control gains indestructible until end of turn.", game.AnswerSurvive},
		{"Serra Advocate", "{T}: Target attacking or blocking creature gets +2/+2 until end of turn.", game.AnswerPump},
		{"Shambling Shell", "Sacrifice this creature: Put a +1/+1 counter on target creature.", game.AnswerPump},
		{"Shidako, Broodmistress", "{G}, Sacrifice a creature: Target creature gets +3/+3 until end of turn.", game.AnswerPump | game.AnswerSacOutlet},
		{"Shieldmage Elder", "Tap two untapped Clerics you control: Prevent all damage target creature would deal this turn.", game.AnswerPrevent},
		{"Shieldmage Elder", "Tap two untapped Wizards you control: Prevent all damage target spell would deal this turn.", game.AnswerPrevent},
		{"Slayers' Stronghold", "{R}{W}, {T}: Target creature gets +2/+0 and gains vigilance and haste until end of turn.", game.AnswerPump | game.AnswerCombatGrant},
		{"Songstitcher", "{1}{W}: Prevent all combat damage that would be dealt this turn by target attacking creature with flying.", game.AnswerPrevent},
		{"Soratami Cloud Chariot", "{2}: Prevent all combat damage that would be dealt to and dealt by target creature you control this turn.", game.AnswerPrevent},
		{"Spearbreaker Behemoth", "{1}: Target creature with power 5 or greater gains indestructible until end of turn.", game.AnswerSurvive},
		{"Spirit Shield", "{2}, {T}: Target creature gets +0/+2 for as long as this artifact remains tapped.", game.AnswerPump},
		{"Stegron the Dinosaur Man", "Dinosaur Formula — {1}{R}, Discard this card: Until end of turn, target creature you control gets +3/+1 and becomes a Dinosaur in addition to its other types.", game.AnswerPump},
		{"Stern Marshal", "{T}: Target creature gets +2/+2 until end of turn. Activate only during your turn, before attackers are declared.", game.AnswerPump},
		{"Stonewise Fortifier", "{4}{W}: Prevent all damage that would be dealt to this creature by target creature this turn.", game.AnswerPrevent},
		{"Streambed Aquitects", "{T}: Target Merfolk creature gets +1/+1 and gains islandwalk until end of turn. (It can't be blocked as long as defending player controls an Island.)", game.AnswerPump | game.AnswerCombatGrant},
		{"Swarmyard", "{T}: Regenerate target Insect, Rat, Spider, or Squirrel.", game.AnswerSurvive},
		{"Sword of the Chosen", "{T}: Target legendary creature gets +2/+2 until end of turn.", game.AnswerPump},
		{"Sylvan Safekeeper", "Sacrifice a land: Target creature you control gains shroud until end of turn.", game.AnswerProtect},
		{"Tangle Tumbler", "{3}, {T}: Put a +1/+1 counter on target creature.", game.AnswerPump},
		{"Tawnos's Weaponry", "{2}, {T}: Target creature gets +1/+1 for as long as this artifact remains tapped.", game.AnswerPump},
		{"The Pride of Hull Clade", "{2}{U}{U}: Until end of turn, target creature you control gets +1/+0, gains \"Whenever this creature deals combat damage to a player, draw cards equal to its toughness,\" and can attack as though it didn't have defender.", game.AnswerPump | game.AnswerCombatGrant},
		{"The Seedcore", "Corrupted — {T}: Target 1/1 creature gets +2/+1 until end of turn. Activate only if an opponent has three or more poison counters.", game.AnswerPump},
		{"Timberwatch Elf", "{T}: Target creature gets +X/+X until end of turn, where X is the number of Elves on the battlefield.", game.AnswerPump},
		{"Tovolar, the Midnight Scourge", "{X}{R}{G}: Target Wolf or Werewolf you control gets +X/+0 and gains trample until end of turn.", game.AnswerPump | game.AnswerCombatGrant},
		{"Tyrite Sanctum", "{4}, {T}, Sacrifice this land: Put an indestructible counter on target God.", game.AnswerSurvive},
		{"Unlikely Alliance", "{1}{W}: Target nonattacking, nonblocking creature gets +0/+2 until end of turn.", game.AnswerPump},
		{"Untamed Pup", "{3}{G}: Put a +1/+1 counter on target creature.", game.AnswerPump},
		{"Venom, Deadly Devourer", "{3}: Exile target creature card from a graveyard. When you do, put X +1/+1 counters on target Symbiote, where X is the exiled card's toughness.", game.AnswerPump | game.AnswerGraveyard},
		{"Viper, Cruel Conspirator", "{B}: Target creature that's attacking alone gets +1/+1 until end of turn.", game.AnswerPump},
		{"Welding Jar", "Sacrifice this artifact: Regenerate target artifact.", game.AnswerSurvive},
		{"Wellgabber Apothecary", "{1}{W}: Prevent all damage that would be dealt to target tapped Merfolk or Kithkin creature this turn.", game.AnswerPrevent},
		{"Wildheart Invoker", "{8}: Target creature gets +5/+5 and gains trample until end of turn.", game.AnswerPump | game.AnswerCombatGrant},
		{"Wishmonger", "{2}: Target creature gains protection from the color of its controller's choice until end of turn. Any player may activate this ability.", game.AnswerProtect},
		{"Wojek Apothecary", "{T}: Prevent the next 1 damage that would be dealt to target creature and each other creature that shares a color with it this turn.", game.AnswerPrevent},
		{"Wyluli Wolf", "{T}: Target creature gets +1/+1 until end of turn.", game.AnswerPump},
		{"Xun Yu, Wei Advisor", "{T}: Target creature you control gets +2/+0 until end of turn. Activate only during your turn, before attackers are declared.", game.AnswerPump},
		{"Yavimaya Hollow", "{G}, {T}: Regenerate target creature.", game.AnswerSurvive},
		{"Yoshimaru, Beloved Companion", "{6}: Put a +1/+1 counter on target legendary creature.", game.AnswerPump},
		{"Yoshimaru, Scrappy Stray", "{6}: Put a +1/+1 counter on target nonlegendary creature.", game.AnswerPump},
		{"Zelyon Sword", "{3}, {T}: Target creature gets +2/+0 for as long as this artifact remains tapped.", game.AnswerPump},
		{"Zhalfirin Commander", "{1}{W}{W}: Target Knight creature gets +1/+1 until end of turn.", game.AnswerPump},
		// Declared by the shield and redirection helpers (preventsDamage),
		// with alsoAnswers where the cost or the redirection adds a kind.
		{"Charm Peddler", "{W}, {T}, Discard a card: The next time a source of your choice would deal damage to target creature this turn, prevent that damage.", game.AnswerPrevent},
		{"Circle of Despair", "{1}, Sacrifice a creature: The next time a source of your choice would deal damage to any target this turn, prevent that damage.", game.AnswerPrevent | game.AnswerSacOutlet},
		{"Daughter of Autumn", "{W}: The next 1 damage that would be dealt to target white creature this turn is dealt to Daughter of Autumn instead.", game.AnswerPrevent},
		{"General's Regalia", "{3}: The next time a source of your choice would deal damage to you this turn, that damage is dealt to target creature you control instead.", game.AnswerPrevent},
		{"Glarecaster", "{5}{W}: The next time damage would be dealt to this creature and/or you this turn, that damage is dealt to any target instead.", game.AnswerPrevent | game.AnswerRemove},
		{"Jade Monolith", "{1}: The next time a source of your choice would deal damage to target creature this turn, that source deals that damage to you instead.", game.AnswerPrevent},
		{"Lancers en-Kor", "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead.", game.AnswerPrevent},
		{"Martyr's Cause", "Sacrifice a creature: The next time a source of your choice would deal damage to any target this turn, prevent that damage.", game.AnswerPrevent | game.AnswerSacOutlet},
		{"Mirrorwood Treefolk", "{2}{R}{W}: The next time damage would be dealt to this creature this turn, that damage is dealt to any target instead.", game.AnswerPrevent | game.AnswerRemove},
		{"Nomads en-Kor", "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead.", game.AnswerPrevent},
		{"Nova Pentacle", "{3}, {T}: The next time a source of your choice would deal damage to you this turn, that damage is dealt to target creature of an opponent's choice instead.", game.AnswerPrevent},
		{"Oracle's Attendants", "{T}: All damage that would be dealt to target creature this turn by a source of your choice is dealt to this creature instead.", game.AnswerPrevent},
		{"Outrider en-Kor", "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead.", game.AnswerPrevent},
		{"Razia, Boros Archangel", "{T}: The next 3 damage that would be dealt to target creature you control this turn is dealt to another target creature instead.", game.AnswerPrevent | game.AnswerRemove},
		{"Sanctum Guardian", "Sacrifice this creature: The next time a source of your choice would deal damage to any target this turn, prevent that damage.", game.AnswerPrevent},
		{"Shaman en-Kor", "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead.", game.AnswerPrevent},
		{"Shaman en-Kor", "{1}{W}: The next time a source of your choice would deal damage to target creature this turn, that damage is dealt to this creature instead.", game.AnswerPrevent},
		{"Soltari Guerrillas", "{0}: The next time this creature would deal combat damage to an opponent this turn, it deals that damage to target creature instead.", game.AnswerPrevent | game.AnswerRemove},
		{"Spirit en-Kor", "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead.", game.AnswerPrevent},
		{"Vassal's Duty", "{1}: The next 1 damage that would be dealt to target legendary creature you control this turn is dealt to you instead.", game.AnswerPrevent},
		{"Ward of Piety", "{1}{W}: The next 1 damage that would be dealt to enchanted creature this turn is dealt to any target instead.", game.AnswerPrevent | game.AnswerRemove},
		{"Warrior en-Kor", "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead.", game.AnswerPrevent},
		{"Zealous Inquisitor", "{1}{W}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature instead.", game.AnswerPrevent | game.AnswerRemove},
		{"Zhalfirin Crusader", "{1}{W}: The next 1 damage that would be dealt to this creature this turn is dealt to any target instead.", game.AnswerPrevent | game.AnswerRemove},
	} {
		t.Run(r.card+": "+r.label, func(t *testing.T) {
			spec := specNamed(t, r.card)
			for _, ab := range spec.Activated {
				if ab.Label != r.label {
					continue
				}
				if hasUntargetedAnnouncement(activatedShapeOf(ab)) {
					t.Errorf("can be announced with no target; this batch is the rows that must name one")
				}
				if got := ab.Purpose.Answers; got != r.want {
					t.Errorf("declares %s, want %s", got, r.want)
				}
				return
			}
			t.Errorf("no activated row with that label")
		})
	}
}
