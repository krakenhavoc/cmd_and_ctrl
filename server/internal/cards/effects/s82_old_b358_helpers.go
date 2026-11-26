package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB358LifeBelowStarting is "your life total is less than your starting
// life total" (Resolute Archangel), the table's starting total (ADR 0075).
func S82OldB358LifeBelowStarting(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Life < StartingLifeOf(g)
}

// S82OldB358GainThatMuchLife is "you gain that much life" for a trigger that
// settled its damage batch (Wall of Essence): the controller gains the
// event's settled amount.
func S82OldB358GainThatMuchLife(g *game.Game, item *game.StackItem) error {
	return GainLife{Amount: SettledEventAmount(item)}.Apply(NewContext(g, item))
}

// S82OldB358CreatureCardPutIntoAnOpponentsGraveyard is "a creature card is put
// into an opponent's graveyard from anywhere" (Profane Memento): Vulturous
// Zombie's reading of a card arriving in an opponent's graveyard, narrowed to
// creature cards.
func S82OldB358CreatureCardPutIntoAnOpponentsGraveyard(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !S82OldB256CardPutIntoAnOpponentsGraveyard(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

// S82OldB358NonlandCardsNamedInAllZones is the cards a "choose a nonland card
// name. Search target player's graveyard, hand, and library for any number of
// cards with that name and exile them" finds (Stain the Mind): every card of
// that player's with the name (a card has it when any face does, CR 201.2b)
// that is not a land.
func S82OldB358NonlandCardsNamedInAllZones(p *game.Player, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, z := range []*game.Zone{p.Graveyard, p.Hand, p.Library} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			if game.CardNameMatches(c, name) && !c.IsLand() {
				out = append(out, c.InstanceID)
			}
		}
	}
	return out
}

// S82OldB358PlayerOrPlaneswalkerAndTheirCreature is the two clauses of "target
// player or planeswalker and … up to one target creature that player or that
// planeswalker's controller controls" (Soul of Shandalar, Chandra,
// Pyromaster): the creature is judged against the controller of the first
// target (ControlledByTargetsController) when it is chosen and again as the
// ability resolves.
func S82OldB358PlayerOrPlaneswalkerAndTheirCreature() *game.TargetSpec {
	return Clauses(
		TargetPlayerOrPlaneswalker(),
		ControlledByTargetsController(
			TargetCreature("up to one target creature that player or that planeswalker's controller controls").WithCount(0, 1), 0),
	)
}

// S82OldB358DamageBothTargets deals n damage from the source to each of the two
// clauses' targets that is still legal, as one damage instruction (CR 120).
func S82OldB358DamageBothTargets(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		var targets []uuid.UUID
		for slot := 0; slot < 2; slot++ {
			if t, ok := ctx.ClauseTarget(slot); ok {
				targets = append(targets, t.ID)
			}
		}
		if len(targets) == 0 {
			return nil
		}
		return g.DealDamageEachKeyedForEffect(item.SourceCardID, targets, n, game.KeyedTail{})
	}
}

// S82OldB358DrawACardForEachColorAmongYourPermanents is "draw a card for each
// color among permanents you control" (Soul of Ravnica).
func S82OldB358DrawACardForEachColorAmongYourPermanents(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := len(ColorsAmongPermanentsYouControl(g, ctx.Controller()))
	if n == 0 {
		return nil
	}
	return DrawCards{N: n}.Apply(ctx)
}

// S82OldB358BecomeElementalLand is "target land becomes a 4/4 Elemental
// creature with trample. It's still a land" (Nissa, Worldwaker) with no
// duration (CR 611.2a): Creature and Elemental are added to what the land is
// (layer 4), trample is granted (layer 6) and the base power and toughness are
// 4/4 (layer 7b). It follows the land only while it is the same object
// (CR 400.7).
func S82OldB358BecomeElementalLand(ctx *Context, land uuid.UUID, label string) error {
	return ScopedEffectFor{
		Target: land,
		Mods: append([]game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
			game.AddKeywordsMod("trample"),
		}, game.SetBasePTMods(4, 4)...),
		Duration: game.IndefiniteDuration(),
		Label:    label,
	}.Apply(ctx)
}
