package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hostileHostelOracle is shared by both faces: the front registers under
// it, the back under hostileHostelOracle + "#1".
const hostileHostelOracle = "1b340f71-502f-48e9-85ed-9af62f356115"

// hostileHostelSoul is the counter kind the Hostel collects.
const hostileHostelSoul = "soul"

// Hostile Hostel // Creeping Inn — a transforming land (Midnight Hunt):
//
//	Hostile Hostel, Land
//	  "{T}: Add {C}.
//	   {1}, {T}, Sacrifice a creature: Put a soul counter on this land.
//	   Then if there are three or more soul counters on it, remove those
//	   counters, transform it, then untap it. Activate only as a
//	   sorcery."
//
//	Creeping Inn, Artifact Creature — Horror Construct, 3/7
//	  "Whenever this creature attacks, you may exile a creature card from
//	   your graveyard. If you do, each opponent loses X life and you gain
//	   X life, where X is the number of creature cards exiled with this
//	   creature.
//	   {4}: This creature phases out."
//
// The Hostel's ability pays the creature as a cost (CR 118.3) and its
// {T} too, so the land is tapped when the counter goes on; the count is
// read as it resolves, and the third soul counter takes all of them off,
// turns the land over in place (TransformThis, CR 701.27f) and untaps it.
// A Hostel that left in response has no counter to place, and nothing
// transforms.
//
// The Inn's trigger asks which creature card to exile (or none), links it
// to the Inn (CR 607.2a: the object, so a Hostel that was never exiled
// from still counts what its Inn exiled) and counts every creature card
// exiled with it, including the ones from earlier attacks, as it drains.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     hostileHostelOracle,
		Name:         "Hostile Hostel",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{1}, {T}, Sacrifice a creature: Put a soul counter on this land. Then if there are three or more soul counters on it, remove those counters, transform it, then untap it. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}"), TapCost(), SacrificeACreature()),
			SorcerySpeed: true,
			Effect:       hostileHostelSoulCounter,
		}},
	})
	Register(Spec{
		OracleID:     hostileHostelOracle + "#1",
		Name:         "Creeping Inn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Creeping Inn — you may exile a creature card from your graveyard; each opponent loses X life and you gain X life",
				creepingInnOffer),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}: This creature phases out.",
			Cost:    ManaCost("{4}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return PhaseOut{Targets: ctx.withoutNewSourceObject([]uuid.UUID{ctx.Source()})}.Apply(ctx)
			},
		}},
	})
}

// hostileHostelSoulCounter is the Hostel's ability: a soul counter, then,
// with three or more on it, take them off, transform it and untap it.
func hostileHostelSoulCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CounterOnThis{Kind: hostileHostelSoul, N: 1}).Apply(ctx); err != nil {
		return err
	}
	src := ctx.Source()
	if !onBattlefield(g, src) {
		return nil
	}
	c, ok := g.LookupCardForEffect(src)
	if !ok || c.Counters[hostileHostelSoul] < 3 {
		return nil
	}
	if _, err := socB9RemoveCounters(ctx, src, hostileHostelSoul, c.Counters[hostileHostelSoul]); err != nil {
		return err
	}
	if err := (TransformThis{}).Apply(ctx); err != nil {
		return err
	}
	return UntapTarget{Target: src}.Apply(ctx)
}

// creepingInnOffer is the attack trigger: offer a creature card of the
// controller's graveyard to exile, or none.
func creepingInnOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var cards []uuid.UUID
	if p := g.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				cards = append(cards, c.InstanceID)
			}
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Creeping Inn — you may exile a creature card from your graveyard",
		Cards:    cards,
		Min:      0,
		Max:      1,
		ThenKey:  creepingInnExiled,
		LiveZone: game.ZoneGraveyard,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// creepingInnExiled exiles the chosen card linked to the Inn, then, if it
// reached exile, drains each opponent for the number of creature cards
// exiled with the Inn. An on-disk key.
var creepingInnExiled = askedPick("choose-cards/creeping-inn-exile", func(ctx *Context, r game.CardsChosen) error {
	g := ctx.Game
	if len(r.Picked) == 0 {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(r.Source)
	if !ok {
		return nil
	}
	if err := g.ExileLinkedForEffect(ref, r.Picked[0]); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(r.Picked[0]); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	x := 0
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() {
			x++
		}
	}
	if x == 0 {
		return nil
	}
	for _, opp := range ctx.Opponents() {
		if err := g.ChangePlayerLifeForEffect(r.Source, opp, -x); err != nil {
			return err
		}
	}
	return g.ChangePlayerLifeForEffect(r.Source, ctx.Controller(), x)
})
