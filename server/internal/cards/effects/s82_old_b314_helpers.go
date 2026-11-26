package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b314_helpers.go — shared pieces of OLD batch 314, Battle for
// Zendikar (S82, #5547).

// S82OldB314BlockedByColorless is Hedron Blade's condition: "Whenever
// equipped creature becomes blocked by one or more colorless creatures".
// EventBecomesBlocked is emitted once per blocked attacker at the block
// declaration's lock-in (CR 506.4), after the blockers are recorded, so the
// trigger is the "one or more" and fires once however many block.
func S82OldB314BlockedByColorless(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesBlocked || !source.IsAttachedTo(ev.CardID) {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.IsColorless() && c.IsBlockingAttacker(ev.CardID) {
			return true
		}
	}
	return false
}

// S82OldB314Ping is Pathway Arrows' granted ability: "This creature deals 1
// damage to target creature. If a colorless creature is dealt damage this
// way, tap it." The damage is the creature's (the ability's source), and the
// tap is a keyed tail told whether damage was really dealt (CR 120.3).
func S82OldB314Ping(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.DealDamageToCreatureKeyedForEffect(item.SourceCardID, t.ID, 1,
			game.KeyedTail{Ref: s82OldB314PingDealt, Player: item.Controller, Source: item.SourceCardID})
	}
	return nil
}

// s82OldB314PingDealt taps the creature that was dealt damage if it is
// colorless. An on-disk key.
var s82OldB314PingDealt = game.RegisterTail("then/s82-old-b314-pathway-arrows-tap", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.IDs) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(o.IDs[0])
	if !ok || !c.IsCreature() || !c.IsColorless() {
		return nil
	}
	ctx, _ := AskedContext(g, nil, o.Tail.Player, o.Tail.Source, nil)
	return TapTarget{Target: o.IDs[0]}.Apply(ctx)
})

// S82OldB314SlabHammer is Slab Hammer's "you may return a land you control
// to its owner's hand. If you do, the creature gets +2/+2 until end of
// turn": the optional pick, the bounce, then the pump on the attacker the
// trigger names.
func S82OldB314SlabHammer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	attacker := ctx.Trigger().Event.CardID
	return ChoosePermanents{
		Question:   "Slab Hammer — you may return a land you control to its owner's hand",
		Candidates: OldB093OptionalLandYouControl,
		Keyed:      s82OldB314SlabHammerPick,
		Carry:      []uuid.UUID{attacker},
	}.Apply(ctx)
}

var s82OldB314SlabHammerPick = RegisterPermanentsPick("prompt-run/s82-old-b314-slab-hammer-return-land", PermanentsDef{
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		if len(picked.Cards()) == 0 || len(r.Carry) == 0 {
			return nil
		}
		land := picked.Cards()[0]
		if err := (BounceToHand{Target: land}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		// "If you do": the land must really have left the battlefield.
		if z := ctx.Game.FindCardZoneForEffect(land); z != nil && z.Kind == game.ZoneBattlefield {
			return nil
		}
		return BoostUntilEOT{Target: r.Carry[0], Power: 2, Toughness: 2, Label: "Slab Hammer — +2/+2"}.Apply(ctx)
	},
})

// S82OldB314RecklessCohortMustAttack is "This creature attacks each combat if
// able unless you control another Ally": the requirement is switched off
// while another Ally is under the controller's control. "Ally" is a word of
// the card's text (CR 612.2).
func S82OldB314RecklessCohortMustAttack() game.StaticAbility {
	s := AttacksEachCombatWhere(func(target *game.Card, g *game.Game, source *game.Card) bool {
		if !SelfOnly(target, g, source) {
			return false
		}
		ally := source.TextSubtype("Ally")
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == source.Controller && c.InstanceID != source.InstanceID && c.HasSubtype(ally) {
				return false
			}
		}
		return true
	})
	s.Label = "Attacks each combat if able unless you control another Ally."
	return s
}

// S82OldB314AlignedHedronNetwork is "When this artifact enters, exile all
// creatures with power 5 or greater until this artifact leaves the
// battlefield": each is exiled under the one record (CR 610.3), power read
// through the layered view as the trigger resolves.
func S82OldB314AlignedHedronNetwork(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.CurrentPower() >= 5 {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (ExileUntil{
			Target:     id,
			ThisLeaves: true,
			Label:      "Aligned Hedron Network — returns when Aligned Hedron Network leaves the battlefield",
		}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB314BlightHerderLook is Blight Herder's cast trigger body: "you may
// put two cards your opponents own from exile into their owners'
// graveyards. If you do, create three 1/1 colorless Eldrazi Scion creature
// tokens." All or nothing: the pick is exactly two cards or none.
func S82OldB314BlightHerderLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	me := ctx.Controller()
	var cands []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Owner != me && !c.IsToken() {
			cands = append(cands, c.InstanceID)
		}
	}
	if len(cands) < 2 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  me,
		Source:   ctx.Source(),
		Question: "Blight Herder — you may put two cards your opponents own from exile into their owners' graveyards",
		Cards:    cands,
		Min:      0,
		Max:      2,
		LiveZone: game.ZoneExile,
		ThenKey:  s82OldB314BlightHerderPick,
	})
	return nil
}

var s82OldB314BlightHerderPick = game.RegisterChooseCardsThenRule("choose-cards/s82-old-b314-blight-herder",
	func(picked []game.Card) bool { return len(picked) == 0 || len(picked) == 2 },
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) != 2 {
			return nil
		}
		// "If you do" waits for the move: a replacement's prompt on either card
		// pauses it, and the Scions are made once both are in a graveyard.
		return g.PutCardsIntoGraveyardKeyedForEffect(r.Picked,
			game.KeyedTail{Ref: s82OldB314BlightHerderMoved, Player: r.Chooser, Source: r.Source})
	})

// s82OldB314BlightHerderMoved creates the Scions once both cards reached a
// graveyard. An on-disk key.
var s82OldB314BlightHerderMoved = game.RegisterTail("then/s82-old-b314-blight-herder-scions", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) != 2 {
		return nil
	}
	ctx, _ := AskedContext(g, nil, o.Tail.Player, o.Tail.Source, nil)
	return CreateToken{Template: B28EldraziScionToken(), N: 3}.Apply(ctx)
})

// S82OldB314KioraTake is Kiora, Master of the Depths's −2: the four cards are
// revealed, the player may put a creature card and/or a land card from among
// them into their hand, and the rest go into the graveyard.
var S82OldB314KioraTake = RegisterTake("choose-cards/s82-old-b314-kiora-take",
	TakeDef{
		Slots: []TakeSlot{
			{Label: "a creature card", Match: Creature(), Max: 1},
			{Label: "a land card", Match: Land(), Max: 1},
		},
		Then: TakeRestIntoGraveyard,
	})

// S82OldB314KioraMinusTwo is the −2 itself.
func S82OldB314KioraMinusTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player:   player,
		Cards:    g.RevealTopOfLibraryForEffect(player, ctx.Source(), 4, "Kiora, Master of the Depths — reveal the top four cards of your library"),
		Optional: true,
		Label:    "Kiora, Master of the Depths — you may put a creature card and/or a land card from among them into your hand",
		Key:      S82OldB314KioraTake,
	}.Apply(ctx)
}
