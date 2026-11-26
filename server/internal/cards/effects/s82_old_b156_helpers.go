package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b156_helpers.go — shared bodies for the Core Set 2020 creatures of
// S82 set OLD, batch 156 (#5518).

// OldB156DifferentManaCostsInGraveyard counts the different mana costs among
// the nonland cards in a player's graveyard, as the string printed on the card
// ({2}{B} and {1}{B}{B} are different; a card with no mana cost has none to
// count).
func OldB156DifferentManaCostsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	seen := map[string]bool{}
	for _, c := range p.Graveyard.Cards {
		if c.IsLand() || c.ManaCost == "" {
			continue
		}
		seen[c.ManaCost] = true
	}
	return len(seen)
}

// OldB156KnightPump is Knight of the Ebon Legion's {2}{B} ability: this
// creature gets +3/+3 and gains deathtouch until end of turn.
func OldB156KnightPump(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	if err := ThisGetsUntilEndOfTurn(3, 3, "Knight of the Ebon Legion — +3/+3")(g, item); err != nil {
		return err
	}
	return ThisGainsKeywordUntilEndOfTurn("deathtouch", "Knight of the Ebon Legion — deathtouch")(g, item)
}

// OldB156OpponentsCreatures is "creatures your opponents control", the
// AppliesTo of Blightbeetle's static.
func OldB156OpponentsCreatures(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller != source.Controller
}

// oldB156AwakenerEnter puts the card onto the battlefield tapped and attacking
// `target` (CR 508.4) and then schedules the CR 603.7 delayed trigger that
// sacrifices exactly that permanent at the beginning of the next end step.
func oldB156AwakenerEnter(ctx *Context, card, target uuid.UUID) error {
	if err := PutAttacking(ctx.Game, ctx.Controller(), card, PutAttackingForm{From: game.ZoneGraveyard}, target); err != nil {
		return err
	}
	if !OnBattlefield(ctx.Game, card) {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Thunderkin Awakener — sacrifice it",
		Cards: []uuid.UUID{card},
		Body:  SacrificeListedCardsBody,
	}.Apply(ctx)
}

// oldB156AwakenerAnswered is the keyed answer to "what does it attack?": the
// card and then the attack targets ride the prompt as data. An on-disk key:
// never renamed, never reused.
var oldB156AwakenerAnswered = OptionPickThen("option-pick/thunderkin-awakener-put-attacking", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) < 2 {
		return nil
	}
	card, targets := r.Carry[0], r.Carry[1:]
	index := r.Index
	if index < 0 || index >= len(targets) {
		index = 0
	}
	return oldB156AwakenerEnter(ctx, card, targets[index])
})

// OldB156AwakenerReturn is Thunderkin Awakener's attack trigger as it
// resolves: return the target Elemental card from your graveyard to the
// battlefield tapped and attacking whatever its controller chooses (CR 508.4,
// never declared as an attacker), and sacrifice it at the beginning of the
// next end step. The question is a pause in which the card can leave, so the
// entry checks the zone again (CR 400.7).
func OldB156AwakenerReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 || legal[0].Kind != game.TargetCard {
		return nil
	}
	id := legal[0].ID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	refs := g.AttackTargetsForEffect(ctx.Controller())
	if len(refs) == 0 {
		return nil
	}
	if len(refs) == 1 {
		return oldB156AwakenerEnter(ctx, id, refs[0].ID)
	}
	options := make([]game.ChoiceOption, len(refs))
	carry := []uuid.UUID{id}
	for i, r := range refs {
		options[i] = game.ChoiceOption{Label: "Attack " + AttackTargetName(g, r.ID)}
		if r.Kind != game.AttackTargetPlayer {
			options[i].Cards = []uuid.UUID{r.ID}
		}
		carry = append(carry, r.ID)
	}
	return PickOption{
		Question: "Thunderkin Awakener — what does it attack?",
		Options:  options,
		ThenKey:  oldB156AwakenerAnswered,
		Carry:    carry,
	}.Apply(ctx)
}

// OldB156ElementalDied is Overgrowth Elemental's death trigger as it resolves:
// you gain 1 life, and if the creature that died was an Elemental (as it last
// existed, CR 603.10a) put a +1/+1 counter on this creature.
func OldB156ElementalDied(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GainLife{Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	if item.Trigger == nil {
		return nil
	}
	dead, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if !ok || !LeftAsSubtype(item.Trigger.Event, dead, ctx.TextSubtype("Elemental")) {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
