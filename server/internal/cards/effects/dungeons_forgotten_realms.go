package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dungeons_forgotten_realms.go files the three Adventures in the
// Forgotten Realms dungeons (ADR 0203 amendment of 2026-10-19, #4100):
// the dungeons a plain "venture into the dungeon" chooses among (CR
// 309.2a, 701.49a). Each is one DungeonSpec, exactly as the Undercity
// is in dungeons.go; the engine lists every dungeon not marked
// VentureIntoOnly (game.dungeonsForVenture), so filing them is what
// makes a plain venture offer all three.
//
// The room text is the Scryfall oracle text from the 2026-10-05 dump,
// verbatim (oracle IDs 5c446a7f-0301-4343-b0df-146cf2db605b,
// 30d0398c-bf5a-4b30-936f-afdb5a109b4b and
// d2ea2605-0ca0-4782-851d-e706bd0114e4). The dungeons' only 2021
// rulings are that a dungeon starts outside the game and that nothing
// but a venture brings one in, both of which the engine already holds.
//
// Baldur's Gate Wilderness (Alchemy) is a digital-only dungeon, not a
// card a paper player owns, and is not filed.

// The dungeon keys: on-disk identities (Player.Dungeon.Key,
// CompletedDungeons), never renamed.
const (
	LostMineOfPhandelverKey  = "lost-mine-of-phandelver"
	DungeonOfTheMadMageKey   = "dungeon-of-the-mad-mage"
	TombOfAnnihilationKey    = "tomb-of-annihilation"
	tombOfAnnihilationName   = "Tomb of Annihilation"
	dungeonOfTheMadMageName  = "Dungeon of the Mad Mage"
	lostMineOfPhandelverName = "Lost Mine of Phandelver"
)

// Lost Mine of Phandelver:
//
//	Cave Entrance — Scry 1. (Leads to: Goblin Lair, Mine Tunnels)
//	Goblin Lair — Create a 1/1 red Goblin creature token. (Leads to:
//	  Storeroom, Dark Pool)
//	Mine Tunnels — Create a Treasure token. (Leads to: Dark Pool,
//	  Fungi Cavern)
//	Storeroom — Put a +1/+1 counter on target creature. (Leads to:
//	  Temple of Dumathoin)
//	Dark Pool — Each opponent loses 1 life and you gain 1 life. (Leads
//	  to: Temple of Dumathoin)
//	Fungi Cavern — Target creature gets -4/-0 until your next turn.
//	  (Leads to: Temple of Dumathoin)
//	Temple of Dumathoin — Draw a card.
func init() {
	registerDungeon(DungeonSpec{
		Key:  LostMineOfPhandelverKey,
		Name: lostMineOfPhandelverName,
		Rooms: []DungeonRoomSpec{
			{
				Key: "cave-entrance", Name: "Cave Entrance",
				Text:    "Scry 1.",
				LeadsTo: []string{"goblin-lair", "mine-tunnels"},
				Effect:  Do(Scry{N: 1}),
			},
			{
				Key: "goblin-lair", Name: "Goblin Lair",
				Text:    "Create a 1/1 red Goblin creature token.",
				LeadsTo: []string{"storeroom", "dark-pool"},
				Effect:  Do(CreateToken{Template: TokenCard("1/1 red Goblin"), N: 1}),
			},
			{
				Key: "mine-tunnels", Name: "Mine Tunnels",
				Text:    "Create a Treasure token.",
				LeadsTo: []string{"dark-pool", "fungi-cavern"},
				Effect:  undercityStash,
			},
			{
				Key: "storeroom", Name: "Storeroom",
				Text:    "Put a +1/+1 counter on target creature.",
				LeadsTo: []string{"temple-of-dumathoin"},
				Targets: TargetCreature("target creature"),
				Effect:  dungeonCounterOnTarget(1),
			},
			{
				Key: "dark-pool", Name: "Dark Pool",
				Text:    "Each opponent loses 1 life and you gain 1 life.",
				LeadsTo: []string{"temple-of-dumathoin"},
				Effect:  Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}),
			},
			{
				Key: "fungi-cavern", Name: "Fungi Cavern",
				Text:    "Target creature gets -4/-0 until your next turn.",
				LeadsTo: []string{"temple-of-dumathoin"},
				Targets: TargetCreature("target creature"),
				Effect:  lostMineFungiCavern,
			},
			{
				Key: "temple-of-dumathoin", Name: "Temple of Dumathoin",
				Text:   "Draw a card.",
				Effect: undercityArchives,
			},
		},
	})
}

// Dungeon of the Mad Mage:
//
//	Yawning Portal — You gain 1 life. (Leads to: Dungeon Level)
//	Dungeon Level — Scry 1. (Leads to: Goblin Bazaar, Twisted Caverns)
//	Goblin Bazaar — Create a Treasure token. (Leads to: Lost Level)
//	Twisted Caverns — Target creature can't attack until your next
//	  turn. (Leads to: Lost Level)
//	Lost Level — Scry 2. (Leads to: Runestone Caverns, Muiral's
//	  Graveyard)
//	Runestone Caverns — Exile the top two cards of your library. You
//	  may play them. (Leads to: Deep Mines)
//	Muiral's Graveyard — Create two 1/1 black Skeleton creature tokens.
//	  (Leads to: Deep Mines)
//	Deep Mines — Scry 3. (Leads to: Mad Wizard's Lair)
//	Mad Wizard's Lair — Draw three cards and reveal them. You may cast
//	  one of them without paying its mana cost.
func init() {
	registerDungeon(DungeonSpec{
		Key:  DungeonOfTheMadMageKey,
		Name: dungeonOfTheMadMageName,
		Rooms: []DungeonRoomSpec{
			{
				Key: "yawning-portal", Name: "Yawning Portal",
				Text:    "You gain 1 life.",
				LeadsTo: []string{"dungeon-level"},
				Effect:  Do(GainLife{Amount: 1}),
			},
			{
				Key: "dungeon-level", Name: "Dungeon Level",
				Text:    "Scry 1.",
				LeadsTo: []string{"goblin-bazaar", "twisted-caverns"},
				Effect:  Do(Scry{N: 1}),
			},
			{
				Key: "goblin-bazaar", Name: "Goblin Bazaar",
				Text:    "Create a Treasure token.",
				LeadsTo: []string{"lost-level"},
				Effect:  undercityStash,
			},
			{
				Key: "twisted-caverns", Name: "Twisted Caverns",
				Text:    "Target creature can't attack until your next turn.",
				LeadsTo: []string{"lost-level"},
				Targets: TargetCreature("target creature"),
				Effect:  madMageTwistedCaverns,
			},
			{
				Key: "lost-level", Name: "Lost Level",
				Text:    "Scry 2.",
				LeadsTo: []string{"runestone-caverns", "muirals-graveyard"},
				Effect:  undercityLostWell,
			},
			{
				Key: "runestone-caverns", Name: "Runestone Caverns",
				Text:    "Exile the top two cards of your library. You may play them.",
				LeadsTo: []string{"deep-mines"},
				Effect:  madMageRunestoneCaverns,
			},
			{
				Key: "muirals-graveyard", Name: "Muiral's Graveyard",
				Text:    "Create two 1/1 black Skeleton creature tokens.",
				LeadsTo: []string{"deep-mines"},
				Effect:  Do(CreateToken{Template: TokenCard("1/1 black Skeleton"), N: 2}),
			},
			{
				Key: "deep-mines", Name: "Deep Mines",
				Text:    "Scry 3.",
				LeadsTo: []string{"mad-wizards-lair"},
				Effect:  Do(Scry{N: 3}),
			},
			{
				Key: "mad-wizards-lair", Name: "Mad Wizard's Lair",
				Text:   "Draw three cards and reveal them. You may cast one of them without paying its mana cost.",
				Effect: madMageMadWizardsLair,
			},
		},
	})
}

// Tomb of Annihilation:
//
//	Trapped Entry — Each player loses 1 life. (Leads to: Veils of
//	  Fear, Oubliette)
//	Veils of Fear — Each player loses 2 life unless they discard a
//	  card. (Leads to: Sandfall Cell)
//	Sandfall Cell — Each player loses 2 life unless they sacrifice a
//	  creature, artifact, or land of their choice. (Leads to: Cradle of
//	  the Death God)
//	Oubliette — Discard a card and sacrifice a creature, an artifact,
//	  and a land. (Leads to: Cradle of the Death God)
//	Cradle of the Death God — Create The Atropal, a legendary 4/4
//	  black God Horror creature token with deathtouch.
func init() {
	registerDungeon(DungeonSpec{
		Key:  TombOfAnnihilationKey,
		Name: tombOfAnnihilationName,
		Rooms: []DungeonRoomSpec{
			{
				Key: "trapped-entry", Name: "Trapped Entry",
				Text:    "Each player loses 1 life.",
				LeadsTo: []string{"veils-of-fear", "oubliette"},
				Effect:  tombTrappedEntry,
			},
			{
				Key: "veils-of-fear", Name: "Veils of Fear",
				Text:    "Each player loses 2 life unless they discard a card.",
				LeadsTo: []string{"sandfall-cell"},
				Effect:  tombUnlessRoom(tombVeils),
			},
			{
				Key: "sandfall-cell", Name: "Sandfall Cell",
				Text:    "Each player loses 2 life unless they sacrifice a creature, artifact, or land of their choice.",
				LeadsTo: []string{"cradle-of-the-death-god"},
				Effect:  tombUnlessRoom(tombSandfall),
			},
			{
				Key: "oubliette", Name: "Oubliette",
				Text:    "Discard a card and sacrifice a creature, an artifact, and a land.",
				LeadsTo: []string{"cradle-of-the-death-god"},
				Effect:  tombOubliette,
			},
			{
				Key: "cradle-of-the-death-god", Name: "Cradle of the Death God",
				Text:   "Create The Atropal, a legendary 4/4 black God Horror creature token with deathtouch.",
				Effect: Do(CreateToken{Template: TokenCard("4/4 black Legendary God Horror named The Atropal with deathtouch"), N: 1}),
			},
		},
	})
}

// dungeonCounterOnTarget is "Put N +1/+1 counters on target creature"
// (Storeroom; Forge is N = 2).
func dungeonCounterOnTarget(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return AddCounter{Target: id, Kind: "+1/+1", N: n}.Apply(ctx)
	}
}

// lostMineFungiCavern: target creature gets -4/-0 until your next turn.
// A layer 7c modification locked to the creature (CR 611.2c), ending as
// the room's owner's next turn begins.
func lostMineFungiCavern(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.ModifyPTMod(-4, 0)},
		Duration: DurationUntilYourNextTurn(ctx, item.Controller),
		Label:    "Lost Mine of Phandelver — Fungi Cavern: -4/-0 until your next turn",
	}.Apply(ctx)
}

// madMageTwistedCaverns: target creature can't attack until your next
// turn.
func madMageTwistedCaverns(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttack)},
		Duration: DurationUntilYourNextTurn(ctx, item.Controller),
		Label:    "Dungeon of the Mad Mage — Twisted Caverns: can't attack until your next turn",
	}.Apply(ctx)
}

// madMageRunestoneCaverns: exile the top two cards of your library; you
// may play them. The printed text states no duration, and CR 611.2a
// makes an effect with none last for the rest of the game — so the
// permission lasts as long as each card stays in exile (a card that
// leaves exile is a new object, CR 400.7). "Play": a land among them
// may be played as the turn's land drop, and every timing rule still
// applies; the costs are paid.
func madMageRunestoneCaverns(g *game.Game, item *game.StackItem) error {
	return ExileTopWithPermission{
		From:        item.Controller,
		GrantTo:     item.Controller,
		N:           2,
		WhileExiled: true,
	}.Apply(NewContext(g, item))
}

// madMageMadWizardsLair: draw three cards and reveal them; you may cast
// one of them without paying its mana cost.
//
// The draw is three individual draws, each of which may be replaced (CR
// 121.6), so "them" is what the draws actually put into the hand. The
// count of cards the player has drawn this turn is carried on the
// continuation, and the drawn cards are the tail of
// Game.DrawnThisTurn past it once the last draw is done. The free cast
// is the resolving ability's (CR 608.2g): it is offered during the
// resolution through the resolution_cast prompt (offerFreeHandCastOver),
// ignores timing, and a land is never offered (a land is not cast).
// Outside a resolution the pass-closed window is the fallback.
func madMageMadWizardsLair(g *game.Game, item *game.StackItem) error {
	you := item.Controller
	return g.DrawNThenForEffect(you, 3, game.DrawThen{
		Ref:    madWizardsLairAfterDraw,
		Player: you,
		Source: item.SourceCardID,
		N:      len(g.DrawnThisTurn[you]),
	})
}

// madWizardsLairAfterDraw reveals the drawn cards and offers the free
// cast. d.N is how many cards the player had drawn this turn before the
// room's three.
var madWizardsLairAfterDraw = game.RegisterDrawThen("dungeon/mad-wizards-lair", func(g *game.Game, d game.DrawThen) error {
	all := g.DrawnThisTurn[d.Player]
	if d.N < 0 || d.N > len(all) {
		return nil
	}
	var drawn []uuid.UUID
	for _, id := range all[d.N:] {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
			drawn = append(drawn, id)
		}
	}
	if len(drawn) == 0 {
		return nil
	}
	g.RevealForEffect(game.RevealSpec{
		Player: d.Player,
		Source: d.Source,
		Reason: "Dungeon of the Mad Mage — Mad Wizard's Lair: reveal the drawn cards",
		Cards:  drawn,
	})
	var candidates []uuid.UUID
	for _, id := range drawn {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
			candidates = append(candidates, id)
		}
	}
	// The keyed free-cast offer (free_hand_cast.go, #4246), so a table
	// waiting on it is a restore point.
	offerFreeHandCastOver(g, d.Player, d.Source, madWizardsLairFreeCast,
		"Mad Wizard's Lair — cast one of them without paying its mana cost?", candidates)
	return nil
})

// madWizardsLairFreeCast is the room's free-cast offer, keyed (#4246).
var madWizardsLairFreeCast = registerFreeHandCastOffer("choose-cards/mad-wizards-lair-cast", dungeonOfTheMadMageName)

// tombTrappedEntry: each player loses 1 life.
func tombTrappedEntry(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range apnapPlayers(g) {
		if err := g.ChangePlayerLifeForEffect(ctx.Source(), id, -1); err != nil {
			return err
		}
	}
	return nil
}

// tombUnless is one of Tomb of Annihilation's two "each player loses 2
// life unless they …" rooms.
type tombUnless int

const (
	tombVeils    tombUnless = iota // "unless they discard a card"
	tombSandfall                   // "unless they sacrifice a creature, artifact, or land of their choice"
)

// tombUnlessLife is the printed 2.
const tombUnlessLife = 2

// tombSandfallSpec is "a creature, artifact, or land": one permanent
// that is any of the three.
func tombSandfallSpec() *game.TargetSpec {
	return TargetPermanent("a creature, artifact, or land", Or(Creature(), Artifact(), Land()))
}

// tombUnlessRoom is the room's effect: each player in APNAP order (CR
// 101.4) chooses, knowing the choices made before theirs.
func tombUnlessRoom(kind tombUnless) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return tombUnlessStep(NewContext(g, item), kind, apnapPlayers(g))
	}
}

// tombUnlessStep asks the first player of `remaining` and hands the
// rest on. A player who cannot do the "unless" (an empty hand; nothing
// to sacrifice) is asked nothing and loses the life (CR 608.2c: as much
// as possible).
func tombUnlessStep(ctx *Context, kind tombUnless, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		victim, rest := remaining[0], remaining[1:]
		p := ctx.Game.PlayerByIDForEffect(victim)
		if p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		var can bool
		var alt string
		switch kind {
		case tombVeils:
			can = p.Hand != nil && p.Hand.Size() > 0
			alt = "Discard a card"
		default:
			can = len(ctx.Game.EffectSacrificeCandidatesForEffect(victim, tombSandfallSpec())) > 0
			alt = "Sacrifice a creature, artifact, or land"
		}
		if !can {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -tombUnlessLife); err != nil {
				return err
			}
			remaining = rest
			continue
		}
		lose := "Lose " + strconv.Itoa(tombUnlessLife) + " life"
		ref := tombVeilsThen
		room := "Veils of Fear"
		if kind == tombSandfall {
			ref, room = tombSandfallThen, "Sandfall Cell"
		}
		return PickOption{
			Player:   victim,
			Question: "Tomb of Annihilation — " + room + ": " + lose + ", or " + alt + "?",
			Options: []game.ChoiceOption{
				{Label: lose, LifeCost: tombUnlessLife},
				{Label: alt},
			},
			ThenKey: ref,
			Carry:   rest,
		}.Apply(ctx)
	}
	return nil
}

// tombVeilsThen and tombSandfallThen are the keyed answers: Carry is the
// players still to be asked, in APNAP order. Keyed, so a table waiting
// on the question is a restore point. Registered in init because the
// answer asks the next player, which names the key again.
var tombVeilsThen, tombSandfallThen game.OptionPickThen

func init() {
	tombVeilsThen = OptionPickThen("option-pick/tomb-of-annihilation-veils-of-fear", func(ctx *Context, r game.OptionPicked) error {
		return tombUnlessAnswered(ctx, tombVeils, r)
	})
	tombSandfallThen = OptionPickThen("option-pick/tomb-of-annihilation-sandfall-cell", func(ctx *Context, r game.OptionPicked) error {
		return tombUnlessAnswered(ctx, tombSandfall, r)
	})
}

// tombVeilsDiscarded is the next player's question once the discard
// Veils of Fear asked for has landed, keyed (#4264).
var tombVeilsDiscarded = game.RegisterDiscardThen("choose-cards/tomb-of-annihilation-veils-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return tombUnlessStep(keyedContext(g, r.Carry[0], r.Source), tombVeils, r.Carry[1:])
})

// tombUnlessAnswered applies one player's answer and asks the next. A
// hand or board that emptied between the question and the answer
// degrades to the life loss.
func tombUnlessAnswered(ctx *Context, kind tombUnless, r game.OptionPicked) error {
	victim, rest := r.Chooser, append([]uuid.UUID(nil), r.Carry...)
	item := ctx.Item
	next := func(g *game.Game) error { return tombUnlessStep(NewContext(g, item), kind, rest) }
	if r.Index == 1 {
		switch kind {
		case tombVeils:
			if p := ctx.Game.PlayerByIDForEffect(victim); p != nil && p.Hand != nil && p.Hand.Size() > 0 {
				// Keyed (#4264): the controller and the players still to
				// be asked ride the prompt as its Carry.
				return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   victim,
					Source:   ctx.Source(),
					N:        1,
					Question: "Tomb of Annihilation — Veils of Fear: discard a card",
				}, tombVeilsDiscarded, append([]uuid.UUID{ctx.Controller()}, rest...))
			}
		default:
			if len(ctx.Game.EffectSacrificeCandidatesForEffect(victim, tombSandfallSpec())) > 0 {
				return ctx.Game.PlayerSacrificesThenForEffect(ctx.Source(), victim, tombSandfallSpec(),
					"Tomb of Annihilation — Sandfall Cell: sacrifice a creature, artifact, or land", 1,
					func(g *game.Game, _ game.PromptedSacrifices) error { return next(g) })
			}
		}
	}
	if r.Index >= 0 {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -tombUnlessLife); err != nil {
			return err
		}
	}
	return next(ctx.Game)
}

// tombOubliette: discard a card and sacrifice a creature, an artifact,
// and a land. One permanent fills one of the three: the sacrifices are
// asked one after another (creature, then artifact, then land), so an
// artifact creature taken as the creature is gone before the artifact
// is chosen, and a player with nothing of a kind skips it (CR 701.21a).
// The discard is asked beside them; nothing waits on it.
func tombOubliette(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := item.Controller
	if p := g.PlayerByIDForEffect(you); p != nil && p.Hand != nil && p.Hand.Size() > 0 {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   you,
			Source:   ctx.Source(),
			N:        1,
			Question: "Tomb of Annihilation — Oubliette: discard a card",
		})
	}
	return oublietteSacrifice(g, ctx.Source(), you, 0)
}

// oublietteSacrifices are Oubliette's three sacrifices, in the order
// they are asked.
var oublietteSacrifices = []struct {
	label string
	pred  func() CardPredicate
}{
	{"a creature", Creature},
	{"an artifact", Artifact},
	{"a land", Land},
}

// oublietteSacrifice asks for Oubliette's i-th sacrifice and, once it
// has settled, the next.
func oublietteSacrifice(g *game.Game, source, you uuid.UUID, i int) error {
	if i >= len(oublietteSacrifices) {
		return nil
	}
	s := oublietteSacrifices[i]
	return g.PlayerSacrificesThenForEffect(source, you, TargetPermanent(s.label, s.pred()),
		"Tomb of Annihilation — Oubliette: sacrifice "+s.label, 1,
		func(g *game.Game, _ game.PromptedSacrifices) error {
			return oublietteSacrifice(g, source, you, i+1)
		})
}
