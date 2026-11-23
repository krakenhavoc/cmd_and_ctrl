package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b194_helpers.go — shared pieces of S82 GAP batch 194 (#5948).

// GapB194NextNoncreatureSpellCondition is the event condition of "the next
// noncreature spell you cast this turn" (Dark Apostle): an EventCast by the
// trigger's controller of a spell that is not a creature. A creature spell
// cast first neither fires it nor uses it up (CR 603.7b).
var GapB194NextNoncreatureSpellCondition = game.DelayedCondition("cast/you-next-cast-a-noncreature-spell",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && !c.IsCreature()
	})

// GapB194CascadeTheSpellBody is the cascade the next noncreature spell has: the
// delayed trigger fires as the spell is cast, above it on the stack, and
// cascades from the spell's mana value (CR 702.85a), read as the spell was
// cast (X counted, CR 202.3e).
var GapB194CascadeTheSpellBody = game.SimpleDelayedBody("cascade/the-spell-you-just-cast", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cast := ctx.PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	mv, ok := g.LastKnownSpellManaValueForEffect(cast[0])
	if !ok {
		return nil
	}
	return g.CascadeForEffect(item.Controller, item.SourceCardID, mv)
})

// GapB194GrantNextNoncreatureSpellCascade is Dark Apostle's effect: "The next
// noncreature spell you cast this turn has cascade."
func GapB194GrantNextNoncreatureSpellCascade(g *game.Game, item *game.StackItem) error {
	return DelayedOnEvent{
		Label:     "Dark Apostle — the next noncreature spell you cast this turn has cascade",
		On:        []game.EventKind{game.EventCast},
		Condition: GapB194NextNoncreatureSpellCondition,
		Body:      GapB194CascadeTheSpellBody,
	}.Apply(NewContext(g, item))
}

// GapB194ExocrineEnters is "When this creature enters, it deals X damage to
// each player and each other creature." X is the value announced for the
// spell (CR 107.3m), carried on the item, so it is exact even if the Exocrine
// has left; a copy or a cheat onto the battlefield has X = 0. The damage is the
// Exocrine's, as it last existed if it left (CR 608.2h), in one batch.
func GapB194ExocrineEnters(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Key:       label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = source.CastX()
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			x := item.Params.Amount
			if x <= 0 {
				return nil
			}
			self := item.SourceCardID
			return DamageEachMatchingAndEachPlayerThen(NewContext(g, item), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.IsCreature() && c.InstanceID != self
			}, x, RecipientThen{})
		},
	}
}

// GapB194CastFromAnywhereButYourHand is the condition of "whenever you cast a
// spell from anywhere other than your hand".
func GapB194CastFromAnywhereButYourHand(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return B25CastFromNotHand(ev, source.Controller)
}

// GapB194DamageEqualToTheCastSpellsManaValue is Keeper of Secrets' effect: it
// deals damage equal to the spell's mana value (as it was cast, X counted) to
// the target opponent.
func GapB194DamageEqualToTheCastSpellsManaValue(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	mv, ok := g.LastKnownSpellManaValueForEffect(ctx.Trigger().Event.CardID)
	if !ok || mv <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: mv}.Apply(ctx)
	}
	return nil
}

// GapB194SpellsCastThisTurnDamage is Noise Marine's effect: it deals damage
// equal to the number of spells you've cast this turn (the Marine's own cast
// counts) to the target, counted as the trigger resolves.
func GapB194SpellsCastThisTurnDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := g.SpellsCastThisTurn[item.Controller].Total
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}

// GapB194ARedSourceYouControlDealtDamage is the condition of "whenever a red
// source you control deals damage to one or more permanents and/or players"
// (The Red Terror): an EventDealDamage of more than 0 from a source of yours
// that is red, to a player or a permanent. The source is read as it dealt the
// damage, wherever it is (a spell on the stack counts).
func GapB194ARedSourceYouControlDealtDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	if !ok || src.Controller != source.Controller || !src.HasColor(source.TextColor("R")) {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) != nil {
		return true
	}
	return OnBattlefield(g, ev.Target)
}

// GapB194PerSource is the batch key of a "whenever <a source> deals damage to
// one or more …" clause that names the SOURCE: one trigger per source per
// simultaneous damage (CR 603.2c), however many recipients.
func GapB194PerSource(ev game.Event, _ *game.Card, _ *game.Game) string {
	return ev.Source.String()
}

// GapB194AnotherSourceYouControlDealtExactlyOne is the condition of "whenever
// another source you control deals exactly 1 damage to a permanent or player"
// (Ghyrson Starn): one damage event of exactly 1 from a source of yours that
// is not Ghyrson.
func GapB194AnotherSourceYouControlDealtExactlyOne(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount != 1 || ev.Source == uuid.Nil || ev.Source == source.InstanceID {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	if !ok || src.Controller != source.Controller {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) != nil {
		return true
	}
	return OnBattlefield(g, ev.Target)
}

// GapB194DealTwoToTheDamaged is Ghyrson Starn's effect: he deals 2 damage to
// that permanent or player, the one the triggering event named. A permanent
// that has left by then takes nothing (CR 608.2b).
func GapB194DealTwoToTheDamaged(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := ctx.Trigger().Event.Target
	if target == uuid.Nil {
		return nil
	}
	if g.PlayerByIDForEffect(target) == nil && !OnBattlefield(g, target) {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: target, Amount: 2}.Apply(ctx)
}

// GapB194CreatureTokensYouControl counts the creature tokens a player
// controls, as a cost query reads them (Magnus the Red).
func GapB194CreatureTokensYouControl(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	n := 0
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller == q.Controller && c.IsCreature() && IsToken(c) {
			n++
		}
	}
	return n
}

// GapB194WarpVortexFlipped is Mutalith Vortex Beast's continuation: one called
// flip per opponent, each in the order the opponents were carried. For each
// flip you win, you draw a card; for each flip you lose, the Beast deals 3
// damage to that player (as it last existed, CR 608.2h).
var GapB194WarpVortexFlipped = game.RegisterCoinFlipThen("coin-flip/mutalith-vortex-beast-warp-vortex", func(g *game.Game, r game.CoinFlipped) error {
	ctx := NewContext(g, &game.StackItem{Controller: r.Flipper, SourceCardID: r.Source})
	for i, won := range r.Result.Won {
		if won {
			if err := (DrawCards{Player: r.Flipper, N: 1}).Apply(ctx); err != nil {
				return err
			}
			continue
		}
		if i >= len(r.Carry) {
			continue
		}
		if err := (DealDamage{Source: r.Source, Target: r.Carry[i], Amount: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// GapB194WarpVortex is "When this creature enters, flip a coin for each
// opponent you have." The opponents still in the game are carried on the call.
func GapB194WarpVortex(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opps := ctx.Opponents()
	if len(opps) == 0 {
		return nil
	}
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Mutalith Vortex Beast — call the coin flips",
		Coins:    len(opps),
		ThenKey:  GapB194WarpVortexFlipped,
		Carry:    opps,
	})
	return nil
}

// GapB194BlueHorrorToken is Pink Horror's Blue Horror: a 2/2 blue and red
// Demon Horror with "Whenever you cast an instant or sorcery spell, this token
// deals 1 damage to any target." The damage's source is the token, so the
// ability lives on the token (a row of token_catalog.go's TokenTemplates).
func GapB194BlueHorrorToken() game.Card { return TokenFromCatalog(printedGapB194BlueHorrorToken) }

// printedGapB194BlueHorrorToken is the token as PRINTED. Its slug is an
// on-disk identity and is never renamed.
func printedGapB194BlueHorrorToken() TokenTemplate {
	return TokenTemplate{
		Slug: "blue-horror",
		Card: game.Card{
			Name:      "Blue Horror",
			TypeLine:  "Token Creature — Demon Horror",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"U", "R"},
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(Or(Instant(), Sorcery()),
				"Blue Horror — 1 damage to any target", DamageToFirstTarget(1)),
				TargetAny()),
		},
		Text: "Whenever you cast an instant or sorcery spell, this token deals 1 damage to any target.",
	}
}
