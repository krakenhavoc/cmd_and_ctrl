package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b704_helpers.go — shared shapes of S82 OLD batch 704 (Urza's Destiny
// damage cards, #5621): Cinder Seer and Scent of Cinder's reveal, Landslide's
// sacrifice, Goblin Festival's coin flip, and the small damage triggers.

// S82OldB704FirstLegalDamageTarget is the first still-legal target of an
// ability or spell with a single target clause. With none, a spell or ability
// whose only target is illegal does not resolve (CR 608.2b).
func S82OldB704FirstLegalDamageTarget(ctx *Context) (game.TargetRef, bool) {
	for _, t := range ctx.LegalTargets() {
		return t, true
	}
	return game.TargetRef{}, false
}

// S82OldB704RevealRedForDamage is "Reveal any number of red cards in your
// hand. <Source> deals X damage to any target, where X is the number of cards
// revealed this way." (Cinder Seer, Scent of Cinder). The cards are chosen as
// it resolves, revealed, and the damage follows; with no red card in hand X is 0
// and nothing happens.
func S82OldB704RevealRedForDamage(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if _, ok := S82OldB704FirstLegalDamageTarget(ctx); !ok {
			return nil
		}
		p := g.PlayerByIDForEffect(ctx.Controller())
		if p == nil || p.Hand == nil {
			return nil
		}
		var red []uuid.UUID
		for _, c := range p.Hand.Cards {
			if c.HasColorWord("R") {
				red = append(red, c.InstanceID)
			}
		}
		if len(red) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: name + " — reveal any number of red cards in your hand",
			Cards:    red,
			Min:      0,
			Max:      len(red),
			LiveZone: game.ZoneHand,
			ThenKey:  s82OldB704RedRevealed,
			Asked:    game.NewAskedItem(item),
		})
		return nil
	}
}

// s82OldB704RedRevealed reveals the chosen cards and deals that much damage. An
// on-disk key: never renamed, never reused.
var s82OldB704RedRevealed = AskedPick("choose-cards/s82-old-b704-reveal-red", func(ctx *Context, r game.CardsChosen) error {
	n := len(r.Picked)
	target, ok := S82OldB704FirstLegalDamageTarget(ctx)
	if n == 0 || !ok {
		return nil
	}
	if err := (RevealCards{Cards: r.Picked, Reason: "Reveal any number of red cards in your hand"}).Apply(ctx); err != nil {
		return err
	}
	if ref, found := ctx.SourceRef(); found {
		return DealDamage{SourceObject: &ref, Target: target.ID, Amount: n}.Apply(ctx)
	}
	return DealDamage{Source: ctx.Source(), Target: target.ID, Amount: n}.Apply(ctx)
})

// S82OldB704Landslide is "Sacrifice any number of Mountains. Landslide deals
// that much damage to target player or planeswalker." With the target gone the
// spell does not resolve, so nothing is sacrificed (CR 608.2b); otherwise the
// chosen Mountains are sacrificed together and the damage is the number that
// were.
func S82OldB704Landslide(_ *game.StackItem, ctx *Context) error {
	if _, ok := S82OldB704FirstLegalDamageTarget(ctx); !ok {
		return nil
	}
	return ChoosePermanents{
		Question: "Landslide — sacrifice any number of Mountains",
		Keyed:    s82OldB704LandslideSacrificed,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.HasSubtypeWord("Mountain") {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 0
		},
	}.Apply(ctx)
}

var s82OldB704LandslideSacrificed = RegisterPermanentsPick("prompt-run/s82-old-b704-landslide-sacrifice", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		n := picked.Count()
		target, ok := S82OldB704FirstLegalDamageTarget(ctx)
		if n == 0 || !ok {
			return nil
		}
		return DealDamage{Source: ctx.Source(), Target: target.ID, Amount: n}.Apply(ctx)
	},
})

// S82OldB704GoblinFestival is "This enchantment deals 1 damage to any target.
// Flip a coin. If you lose the flip, choose one of your opponents. That player
// gains control of this enchantment."
func S82OldB704GoblinFestival(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := S82OldB704FirstLegalDamageTarget(ctx)
	if !ok {
		return nil
	}
	if err := (DealDamage{Source: item.SourceCardID, Target: target.ID, Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Goblin Festival — call the coin flip",
		ThenKey:  s82OldB704FestivalFlipped,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB704FestivalFlipped: losing the flip asks the controller for one of
// their opponents. On-disk keys: never renamed, never reused.
var s82OldB704FestivalFlipped = CoinFlipThen("coin-flip/s82-old-b704-goblin-festival", func(ctx *Context, r game.CoinFlipped) error {
	if WonFirstFlip(r) {
		return nil
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Goblin Festival — choose one of your opponents; they gain control of this enchantment",
		ThenKey:  s82OldB704FestivalOpponent,
	}.Apply(ctx)
})

var s82OldB704FestivalOpponent = ChoosePlayerThen("option-pick/s82-old-b704-goblin-festival-opponent", func(ctx *Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil || !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return GainControl{
		Target:     ctx.Item.SourceCardID,
		Controller: chosen,
		Duration:   game.IndefiniteDuration(),
		Label:      "Goblin Festival — the chosen opponent gains control of it",
	}.Apply(ctx)
})

// S82OldB704AetherStingDamage is "this enchantment deals 1 damage to that
// player": the opponent who cast the creature spell.
func S82OldB704AetherStingDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := TriggeringActor(item)
	ref, ok := ctx.SourceRef()
	if player == uuid.Nil || !ok {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: player, Amount: 1}.Apply(ctx)
}

// S82OldB704ImpatienceDamage is "this enchantment deals 2 damage to that
// player", the player whose end step it is.
func S82OldB704ImpatienceDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := TriggeringActor(item)
	ref, ok := ctx.SourceRef()
	if player == uuid.Nil || !ok {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: player, Amount: 2}.Apply(ctx)
}

// S82OldB704DidntCastASpellThisTurn is the intervening "if that player didn't
// cast a spell this turn" (CR 603.4) for the player whose end step began,
// read from the triggering event as the trigger fires and again as it resolves.
func S82OldB704DidntCastASpellThisTurn(ctx *Context) bool {
	player := TriggeringActor(ctx.Item)
	return player != uuid.Nil && ctx.Game.CastTallyFor(player).Total == 0
}

// S82OldB704CaltropsDamage is "this artifact deals 1 damage to it" for the
// creature that attacked, while it is still that object (CR 400.7).
func S82OldB704CaltropsDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	ref, found := ctx.SourceRef()
	if !ok || info.Left || !found {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: ctx.Trigger().Event.CardID, Amount: 1}.Apply(ctx)
}
