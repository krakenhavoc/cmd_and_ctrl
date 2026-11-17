package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b20_helpers.go — The Brothers' War, S82 batch 20 (#5482).

// s82BroB20DiscardHandThenDrawThree is the yes to "you may discard your
// hand. If you do, draw three cards" (Phyrexian Dragon Engine). A hand with
// no cards in it may still be discarded, so a yes always draws, as it does
// for Vaultguard Trooper.
func s82BroB20DiscardHandThenDrawThree(ctx *Context, r game.ConfirmAnswer) error {
	if _, err := discardWholeHand(ctx.Game, r.Chooser); err != nil {
		return err
	}
	return DrawCards{Player: r.Chooser, N: 3}.Apply(ctx)
}

// s82BroB20DragonEngineDiscard is the on-disk key of the Dragon Engine's
// yes: never renamed, never reused.
var s82BroB20DragonEngineDiscard = MayYes("confirm/phyrexian-dragon-engine-discard-hand", s82BroB20DiscardHandThenDrawThree)

// s82BroB20AttackingCreatures counts the creatures attacking right now,
// whoever controls them: "X is the number of attacking creatures".
func s82BroB20AttackingCreatures(g *game.Game) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget != uuid.Nil && c.IsCreature() {
			n++
		}
	}
	return n
}

// s82BroB20AttackingPartner finds the permanent Mishra, Claimed by Gix
// would meld with: "this creature and a creature named <partner> are
// attacking, and you both own and control them". uuid.Nil when either is
// not attacking, or either is not owned and controlled by `controller`. A
// real meld card is preferred over a copy that only shares the name.
func s82BroB20AttackingPartner(g *game.Game, src, controller uuid.UUID, partner string) uuid.UUID {
	ok := func(c *game.Card) bool {
		return c.Owner == controller && c.Controller == controller && !c.FaceDown &&
			c.AttackingTarget != uuid.Nil && c.IsCreature()
	}
	self, found := g.LookupCardForEffect(src)
	if !found || !onBattlefield(g, src) || !ok(&self) {
		return uuid.Nil
	}
	best := uuid.Nil
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == src || !ok(c) || !game.CardNameMatches(*c, partner) {
			continue
		}
		if c.IsMeldCard() {
			return c.InstanceID
		}
		if best == uuid.Nil {
			best = c.InstanceID
		}
	}
	return best
}

// s82BroB20YouControlALegendaryGreenCreature is Argoth, Sanctum of Nature's
// "unless you control a legendary green creature", read through the layers.
// "Green" is a word of Argoth's text (CR 612.2), so it is read off the source.
func s82BroB20YouControlALegendaryGreenCreature(g *game.Game, src *game.Card) bool {
	green := src.TextColor("G")
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == src.Controller && c.IsCreature() && c.IsLegendary() && c.HasColor(green) {
			return true
		}
	}
	return false
}

// s82BroB20TitaniaMeldReady is the condition of Titania, Voice of Gaea's
// upkeep meld, an intervening "if" (CR 603.4): four or more land cards in
// your graveyard, and you both own and control Titania and a land named
// Argoth, Sanctum of Nature.
type s82BroB20TitaniaMeldReady struct{}

func (s82BroB20TitaniaMeldReady) Holds(ctx *Context) bool {
	if b17LandCardsInGraveyard(ctx.Game, ctx.Controller()) < 4 {
		return false
	}
	src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok {
		return false
	}
	return YouOwnAndControlThisAndANamed("Argoth, Sanctum of Nature", "land")(ctx.Game, &src)
}

func (s82BroB20TitaniaMeldReady) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// s82BroB20SaheeliAfterScry is the rest of Saheeli, Filigree Master's +1,
// run once the scry is done: "You may tap an untapped artifact you control.
// If you do, draw a card." The controller names the artifact, or none.
var s82BroB20SaheeliAfterScry = CardTail("saheeli-filigree-master/after-scry", func(ctx *Context, _ game.TailOutcome) error {
	var untapped []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, And(Artifact(), YouControl())) {
		if !c.Tapped {
			untapped = append(untapped, c.InstanceID)
		}
	}
	if len(untapped) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Saheeli, Filigree Master — you may tap an untapped artifact you control; if you do, draw a card",
		Cards:    untapped,
		Min:      0,
		Max:      1,
		ThenKey:  s82BroB20SaheeliTapArtifact,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// s82BroB20SaheeliTapArtifact is the pick: the chosen artifact is tapped,
// and the card is drawn only if it was tapped (an artifact that tapped or
// left in the meantime gives nothing).
var s82BroB20SaheeliTapArtifact = askedPick("choose-cards/saheeli-filigree-master-tap-artifact", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(r.Picked[0])
	if !ok || !onBattlefield(ctx.Game, c.InstanceID) || c.Tapped || c.Controller != ctx.Controller() || !c.IsArtifact() {
		return nil
	}
	if err := (TapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{N: 1}.Apply(ctx)
})

// s82BroB20ThopterHaste is "they gain haste until end of turn" over the
// Thopters that landed (Saheeli, Filigree Master's −2). An on-disk key.
var s82BroB20ThopterHaste = CardTail("saheeli-filigree-master/thopter-haste", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Saheeli, Filigree Master — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// s82BroB20BrothersWarClash is chapter II of The Brothers' War: "Choose two
// target players. Until your next turn, each creature they control attacks
// the other chosen player each combat if able." It is a rule, not a
// characteristic, so it reaches a creature that comes under a chosen
// player's control later (CR 611.2c): one live record per chosen player,
// scoped to the creatures that player controls, each requiring an attack on
// the other. A player the chapter can no longer affect (a target that became
// illegal, CR 608.2b) gets no requirement, but is still "the other chosen
// player" for the one who remains.
func s82BroB20BrothersWarClash(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var chosen []game.TargetRef
	for _, t := range item.Targets {
		if t.Kind == game.TargetPlayer {
			chosen = append(chosen, t)
		}
	}
	if len(chosen) != 2 {
		return nil
	}
	you := ctx.Controller()
	for i, t := range chosen {
		other := chosen[1-i]
		if !ctx.IsTargetLegal(t) {
			continue
		}
		g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, t.ID,
			[]game.Mod{game.AttackPlayerIfAbleMod(other.ID)},
			DurationUntilYourNextTurn(ctx, you),
			"The Brothers' War — each creature they control attacks the other chosen player each combat if able")
	}
	return nil
}

// s82BroB20BrothersWarBurst is chapter III: "This Saga deals X damage to any
// target and X damage to any other target, where X is the number of
// artifacts you control." X is counted as it resolves; both are one damage
// event from the Saga, and a target that became illegal is dealt nothing
// while the other still is (CR 608.2b).
func s82BroB20BrothersWarBurst(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := b03ArtifactsControlled(g, ctx.Controller())
	if x <= 0 {
		return nil
	}
	return damageEachClauseTarget(x, x)(item, ctx)
}
