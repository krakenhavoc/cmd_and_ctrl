package effectstest

// daynight_b_cards_test.go — the second half of the day/night card list
// (#2586, slice daynight-b). The werewolves go through the importer
// (dnaWerewolf.row) so the per-face daybound / nightbound stamping is on
// the road under test; the rest are pushed catalogue permanents.

const (
	DnbOakshade   = "1b639537-a8fe-4615-b937-afa478fdec0f"
	DnbLiberator  = "9545f126-062f-4410-a362-e16255a128d6"
	DnbStormseek  = "ea5fd21a-c23a-49ee-aab8-0a9618d65c11"
	DnbShady      = "10be1b27-bc9f-4c6e-ac85-f1fa8b2a34d6"
	DnbPainter    = "4ff8c359-430f-434e-ac1e-822abdc28360"
	DnbPhoenix    = "94f14f42-ea66-4432-81cc-667ad62a4497"
	DnbStowaway   = "da412474-1da8-409b-bd7b-87cbfc239fb0"
	DnbRuffian    = "73a3b9a1-37a0-469a-9557-8c118a1ee78f"
	DnbHauler     = "c31e9db3-5d9d-470a-871a-b4b5b0536db5"
	DnbHuntmaster = "18563bc9-6090-4629-b304-89a67d93f635"
	DnbTovolar    = "45d49831-548a-4a0e-9a18-9f7397913895"
	DnbMoonrise   = "df33b298-9e4b-4522-96e5-c8d8d6af2bb4"
	DnbVadrik     = "a9ab17ce-ba07-4ce8-99f9-9c4c633957f7"
	DnbArsonist   = "215dfa88-b130-44df-9cfc-f1f0e4a36f4d"
	DnbPrisoner   = "bdcf0af3-3976-400d-a8b7-15e959e2b255"
	DnbWeaver     = "a64ecae7-0b09-48e1-8108-89442547ffda"
	DnbWolfStrike = "34237f39-1db3-4b3d-b87b-5fd1fbd5462a"
	DnbOutcast    = "ac3d07cd-88c4-4e22-9d86-f92d34f406d4"
)

var DnbWerewolves = []DnaWerewolf{
	{DnbOakshade, "Oakshade Stalker", "Moonlit Ambusher", "Creature — Human Ranger Werewolf", "Creature — Werewolf", "{2}{G}",
		"You may cast this spell as though it had flash if you pay {2} more to cast it.\n" + DnDay,
		DnNight, "3", "3", "6", "3", []string{"G"}, nil, nil, nil},
	{DnbLiberator, "Outland Liberator", "Frenzied Trapbreaker", "Creature — Human Werewolf", "Creature — Werewolf", "{1}{G}",
		"{1}, Sacrifice this creature: Destroy target artifact or enchantment.\n" + DnDay,
		"{1}, Sacrifice this creature: Destroy target artifact or enchantment.\nWhenever this creature attacks, destroy target artifact or enchantment defending player controls.\n" + DnNight,
		"2", "2", "3", "3", []string{"G"}, nil, nil, nil},
	{DnbStormseek, "Reckless Stormseeker", "Storm-Charged Slasher", "Creature — Human Werewolf", "Creature — Werewolf", "{2}{R}",
		"At the beginning of combat on your turn, target creature you control gets +1/+0 and gains haste until end of turn.\n" + DnDay,
		"At the beginning of combat on your turn, target creature you control gets +2/+0 and gains trample and haste until end of turn.\n" + DnNight,
		"2", "3", "3", "4", []string{"R"}, nil, nil, nil},
	{DnbShady, "Shady Traveler", "Stalking Predator", "Creature — Human Werewolf", "Creature — Werewolf", "{2}{B}",
		"Menace (This creature can't be blocked except by two or more creatures.)\n" + DnDay,
		"Menace (This creature can't be blocked except by two or more creatures.)\n" + DnNight,
		"2", "3", "4", "4", []string{"B"}, []string{"Menace"}, []string{"menace"}, []string{"menace"}},
	{DnbPainter, "Spellrune Painter", "Spellrune Howler", "Creature — Human Shaman Werewolf", "Creature — Werewolf", "{2}{R}",
		"Whenever you cast an instant or sorcery spell, this creature gets +1/+1 until end of turn.\n" + DnDay,
		"Whenever you cast an instant or sorcery spell, this creature gets +2/+2 until end of turn.\n" + DnNight,
		"2", "3", "3", "4", []string{"R"}, nil, nil, nil},
	{DnbStowaway, "Suspicious Stowaway", "Seafaring Werewolf", "Creature — Human Rogue Werewolf", "Creature — Werewolf", "{1}{U}",
		"This creature can't be blocked.\nWhenever this creature deals combat damage to a player, draw a card, then discard a card.\n" + DnDay,
		"This creature can't be blocked.\nWhenever this creature deals combat damage to a player, draw a card.\n" + DnNight,
		"1", "1", "2", "1", []string{"U"}, nil, nil, nil},
	{DnbRuffian, "Tavern Ruffian", "Tavern Smasher", "Creature — Human Warrior Werewolf", "Creature — Werewolf", "{3}{R}",
		DnDay, DnNight, "2", "5", "6", "5", []string{"R"}, nil, nil, nil},
	{DnbHauler, "Tireless Hauler", "Dire-Strain Brawler", "Creature — Human Werewolf", "Creature — Werewolf", "{4}{G}",
		"Vigilance\n" + DnDay, "Vigilance\n" + DnNight,
		"4", "5", "6", "6", []string{"G"}, []string{"Vigilance"}, []string{"vigilance"}, []string{"vigilance"}},
	{DnbHuntmaster, "Tovolar's Huntmaster", "Tovolar's Packleader", "Creature — Human Werewolf", "Creature — Werewolf", "{4}{G}{G}",
		"When this creature enters, create two 2/2 green Wolf creature tokens.\n" + DnDay,
		"Whenever this creature enters or attacks, create two 2/2 green Wolf creature tokens.\n{2}{G}{G}: Another target Wolf or Werewolf you control fights target creature you don't control.\n" + DnNight,
		"6", "6", "7", "7", []string{"G"}, nil, nil, nil},
	{DnbTovolar, "Tovolar, Dire Overlord", "Tovolar, the Midnight Scourge", "Legendary Creature — Human Werewolf", "Legendary Creature — Werewolf", "{1}{R}{G}",
		"Whenever a Wolf or Werewolf you control deals combat damage to a player, draw a card.\nAt the beginning of your upkeep, if you control three or more Wolves and/or Werewolves, it becomes night. Then transform any number of Human Werewolves you control.\nDaybound",
		"Whenever a Wolf or Werewolf you control deals combat damage to a player, draw a card.\n{X}{R}{G}: Target Wolf or Werewolf you control gets +X/+0 and gains trample until end of turn.\nNightbound",
		"3", "3", "4", "4", []string{"R", "G"}, nil, nil, nil},
	{DnbArsonist, "Volatile Arsonist", "Dire-Strain Anarchist", "Creature — Human Werewolf", "Creature — Werewolf", "{3}{R}{R}",
		"Menace, haste\nWhenever this creature attacks, it deals 1 damage to each of up to one target creature, up to one target player, and/or up to one target planeswalker.\n" + DnDay,
		"Menace, haste\nWhenever this creature attacks, it deals 2 damage to each of up to one target creature, up to one target player, and/or up to one target planeswalker.\n" + DnNight,
		"4", "4", "5", "5", []string{"R"}, []string{"Menace", "Haste"}, []string{"menace", "haste"}, []string{"menace", "haste"}},
	{DnbPrisoner, "Weary Prisoner", "Wrathful Jailbreaker", "Creature — Human Werewolf", "Creature — Werewolf", "{3}{R}",
		"Defender\n" + DnDay, "This creature attacks each combat if able.\n" + DnNight,
		"2", "6", "6", "6", []string{"R"}, []string{"Defender"}, []string{"defender"}, nil},
	{DnbWeaver, "Weaver of Blossoms", "Blossom-Clad Werewolf", "Creature — Human Werewolf", "Creature — Werewolf", "{2}{G}",
		"{T}: Add one mana of any color.\n" + DnDay, "{T}: Add two mana of any one color.\n" + DnNight,
		"2", "3", "3", "4", []string{"G"}, nil, nil, nil},
	{DnbOutcast, "Wolfkin Outcast", "Wedding Crasher", "Creature — Human Werewolf", "Creature — Werewolf", "{5}{G}",
		"This spell costs {2} less to cast if you control a Wolf or Werewolf.\n" + DnDay,
		"Whenever this creature or another Wolf or Werewolf you control dies, draw a card.\n" + DnNight,
		"5", "4", "6", "5", []string{"G"}, nil, nil, nil},
}

func DnbWerewolf(oracle string) DnaWerewolf {
	for _, w := range DnbWerewolves {
		if w.Oracle == oracle {
			return w
		}
	}
	panic("no such werewolf " + oracle)
}
