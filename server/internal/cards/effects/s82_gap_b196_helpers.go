package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b196_helpers.go — shared pieces of S82 GAP batch 196 (#5948).

// GapB196HealingTearsTargets is the target clause of Celestine's trigger:
// "target creature card with mana value X or less from your graveyard, where
// X is the amount of life you gained this turn". X is read as the trigger is
// put on the stack (and the clause is judged again as it resolves).
func GapB196HealingTearsTargets(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	x := B15LifeGainedThisTurn(g, source.Controller)
	return TargetCardInGraveyard("target creature card with mana value X or less from your graveyard, where X is the amount of life you gained this turn",
		Creature(), YouOwn(), ManaValueLE(x))
}

// GapB196RitesOfBanishment is Grey Knight Paragon's effect: destroy the target
// attacking creature; if it is a Demon, exile it instead.
func GapB196RitesOfBanishment(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	if c, ok := g.LookupCardForEffect(id); ok && c.HasSubtype(ctx.TextSubtype("Demon")) {
		return ExileTarget{Target: id}.Apply(ctx)
	}
	return DestroyTarget{Target: id}.Apply(ctx)
}

// GapB196GravCannon is Space Marine Devastator's effect: destroy up to one
// target artifact or enchantment.
func GapB196GravCannon(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(ctx)
}

// GapB196ConcealedPosition is Space Marine Scout's effect: you may search your
// library for a Plains card, put it onto the battlefield tapped, then shuffle.
func GapB196ConcealedPosition(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	plains := ctx.TextSubtype("Plains")
	return SearchLibrary{
		Player:        item.Controller,
		Source:        item.SourceCardID,
		Predicate:     func(c game.Card) bool { return c.HasSubtype(plains) },
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Optional:      true,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Space Marine Scout — you may search for a Plains card and put it onto the battlefield tapped",
	}.Apply(ctx)
}

// GapB196PutIntoYourGraveyardFromTheBattlefield is the condition of "when this
// creature is put into your graveyard from the battlefield": it died, and its
// owner is the player who controlled it as it left (a stolen one goes to its
// owner's graveyard, which is not "yours" for its last controller).
func GapB196PutIntoYourGraveyardFromTheBattlefield(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !CardDied(ev, source) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == LeftUnderControlOf(ev, c)
}

// GapB196PraesidiumPile is the end of Praesidium Protectiva's exile: if the
// pile (the top cards that reached exile and the Praesidium itself, the first
// carried id, if it still sits in exile) is not empty, it is shuffled and put
// back on top of the library, one random order for the whole pile.
var GapB196PraesidiumPile = CardTail("exile-pile/praesidium-protectiva-back-on-top", func(ctx *Context, o game.TailOutcome) error {
	pile := append([]uuid.UUID(nil), o.IDs...)
	if self := TailCarry0(o); self != uuid.Nil {
		if z := ctx.Game.FindCardZoneForEffect(self); z != nil && z.Kind == game.ZoneExile {
			pile = append(pile, self)
		}
	}
	if len(pile) == 0 {
		return nil
	}
	return ctx.Game.PutOnTopInRandomOrderForEffect(ctx.Controller(), game.ZoneExile, pile)
})

// GapB196PraesidiumProtectiva is the trigger's effect: exile the Triumph and
// the top six cards of your library in a face-down pile; if you do, shuffle
// that pile and put it back on top of your library.
func GapB196PraesidiumProtectiva(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self := item.SourceCardID
	if z := g.FindCardZoneForEffect(self); z != nil && z.Kind == game.ZoneGraveyard {
		if err := (ExileTarget{Target: self}).Apply(ctx); err != nil {
			return err
		}
	}
	return g.ExileTopFaceDownKeyedForEffect(item.Controller, 6, WithCarry(ctx.ItemTail(GapB196PraesidiumPile), self))
}

// GapB196CommandersYouControl is the scope of "commanders you control".
func GapB196CommandersYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCommander && target.Controller == source.Controller
}

// AddTheArtifactTypeOnce is the layer 4 body of "<these> are artifacts in
// addition to their other types" (Mycosynth Lattice, Biotransference): the
// artifact type is appended once.
func AddTheArtifactTypeOnce(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
	for _, t := range c.Types {
		if t == "Artifact" {
			return
		}
	}
	c.Types = append(c.Types, "Artifact")
}

// GapB196NurglesRotDies is the effect of "when enchanted creature dies, return
// this card to its owner's hand and you create a 1/3 black Demon creature token
// named Plaguebearer of Nurgle": the Aura comes back only if it is still the
// card the state-based action put into the graveyard (CR 400.7); the token is
// made either way, for the Aura's controller.
func GapB196NurglesRotDies(g *game.Game, item *game.StackItem) error {
	if err := EndlessEvilReturn(g, item); err != nil {
		return err
	}
	return CreateToken{Template: TokenCard("1/3 black Demon named Plaguebearer of Nurgle"), N: 1}.Apply(NewContext(g, item))
}

// GapB196OutOfTheTombsDraw is the replaced draw of Out of the Tombs: instead of
// drawing from an empty library, return a creature card from your graveyard to
// the battlefield; if you can't, you lose the game. The rest of the draw
// instruction (the parked DrawDone) is paid once the creature has come back.
var GapB196OutOfTheTombsDraw = game.RegisterDrawInstead("out-of-the-tombs", func(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	candidates := GyCardIDs(g, drawer, game.Card.IsCreature)
	if len(candidates) == 0 {
		_, err := g.LoseTheGameForEffect(drawer, source)
		return err
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    drawer,
		FromPlayer: drawer,
		Source:     source,
		Question:   "Out of the Tombs — return a creature card from your graveyard to the battlefield",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		ThenKey:    gapB196OutOfTheTombsReturn,
		Carry:      []uuid.UUID{done.Park(g)},
	})
	return nil
})

// gapB196OutOfTheTombsReturn puts the chosen creature card onto the battlefield
// under its chooser's control, then pays the rest of the draw. An on-disk key.
var gapB196OutOfTheTombsReturn = game.RegisterChooseCardsThen("choose-cards/out-of-the-tombs-return", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard || z.Owner != r.Chooser {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); !ok || !c.IsCreature() {
			continue
		}
		if err := NothingIfGone(g.ReturnFromGraveyardForEffect(id, game.ZoneBattlefield)); err != nil {
			return err
		}
	}
	if len(r.Carry) == 0 {
		return nil
	}
	return game.DrawDoneByID(r.Carry[0]).Run(g)
})

// GapB196OutOfTheTombsReplacement is "If you would draw a card while your
// library has no cards in it, instead return a creature card from your
// graveyard to the battlefield. If you can't, you lose the game." It is
// Laboratory Maniac's window (CR 614.11: it applies even though no card could
// be drawn), for the Tombs' controller's draws.
func GapB196OutOfTheTombsReplacement() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:     []game.EventKind{game.EventDrawCard},
		AppliesTo:   DrawFromAnEmptyLibraryApplies,
		DrawInstead: GapB196OutOfTheTombsDraw,
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Out of the Tombs: return a creature card from your graveyard to the battlefield instead of drawing from an empty library; if you can't, you lose the game",
	}
}

// GapB196OutOfTheTombsUpkeep is "At the beginning of your upkeep, put two eon
// counters on this enchantment, then mill cards equal to the number of eon
// counters on it."
func GapB196OutOfTheTombsUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	if err := (AddCounter{Target: item.SourceCardID, Kind: "eon", N: 2}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	n := c.Counters["eon"]
	if n <= 0 {
		return nil
	}
	return MillCards{Player: item.Controller, N: n}.Apply(ctx)
}

// DrawFromAnEmptyLibraryApplies is the window of "if you would draw a card while
// your library has no cards in it" (Laboratory Maniac's, CR 614.11: it applies
// even though no card could be drawn): a draw of the source's controller whose
// library is empty. The library is read at each draw.
func DrawFromAnEmptyLibraryApplies(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventDraw || ev.DrawCount <= 0 || src == nil || ev.DrawPlayer != src.Controller {
		return false
	}
	p := g.PlayerByIDForEffect(ev.DrawPlayer)
	return p != nil && !p.Eliminated && p.Library.Size() == 0
}
