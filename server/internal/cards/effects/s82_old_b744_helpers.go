package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b744_helpers.go — Exodus, S82 OLD batch 744 (#5626).

// S82OldB744TargetOpponentWithTwoFewerCreatureCards is Keeper of the Dead's
// "target opponent who has at least two fewer creature cards in their
// graveyard than you do as you activate this ability". The graveyards are
// judged as the ability is activated (CR 602.2b).
func S82OldB744TargetOpponentWithTwoFewerCreatureCards() *game.TargetSpec {
	return TargetPlayer("target opponent who has at least two fewer creature cards in their graveyard than you do",
		Opponent(),
		func(g *game.Game, caster uuid.UUID, p *game.Player) bool {
			me := g.PlayerByIDForEffect(caster)
			return me != nil && s82OldB744CreatureCardsInGraveyard(me) >= s82OldB744CreatureCardsInGraveyard(p)+2
		})
}

func s82OldB744CreatureCardsInGraveyard(p *game.Player) int {
	n := 0
	if p == nil || p.Graveyard == nil {
		return 0
	}
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			n++
		}
	}
	return n
}

// S82OldB744DestroyTheTargetCreature destroys the first card target still
// legal as the ability resolves (CR 608.2b), in a row whose first clause is a
// player who is only a condition (Keeper of the Dead).
func S82OldB744DestroyTheTargetCreature(g *game.Game, item *game.StackItem) error {
	return DestroyTheFirstLegalCardTarget(NewContext(g, item))
}

// S82OldB744ExileBlockedCreatures is Wall of Nets' "at end of combat, exile
// all creatures blocked by this creature": the attackers the source is still
// blocking as the end of combat step begins. The exile is linked to the
// trigger's label so the leave trigger can read it back
// (B41ReturnCardsExiledWithToTheBattlefield).
func S82OldB744ExileBlockedCreatures(g *game.Game, item *game.StackItem) error {
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || SourceIsNewObject(g, item) {
		return nil
	}
	blocked := map[uuid.UUID]bool{}
	for _, id := range src.BlockedAttackers() {
		blocked[id] = true
	}
	if len(blocked) == 0 {
		return nil
	}
	return ExileAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return blocked[c.InstanceID] && c.IsCreature() && c.AttackingTarget != uuid.Nil
	}}.Apply(NewContext(g, item))
}

// S82OldB744DauthiWarlordPower is Dauthi Warlord's characteristic-defining
// power (CR 604.3, layer 7a): the number of creatures on the battlefield with
// shadow, itself included, read after layer 6.
func S82OldB744DauthiWarlordPower(g *game.Game) int {
	n := 0
	for i := range g.Battlefield.Cards {
		if p := &g.Battlefield.Cards[i]; p.IsCreature() && game.HasKeyword(p, "shadow") {
			n++
		}
	}
	return n
}

// S82OldB744NonbasicLandsControlledBy counts the nonbasic lands `player`
// controls (Skyshroud War Beast's power and toughness).
func S82OldB744NonbasicLandsControlledBy(g *game.Game, player uuid.UUID) int {
	n := 0
	if player == uuid.Nil {
		return 0
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.IsLand() && !c.HasSupertype("Basic") {
			n++
		}
	}
	return n
}

// S82OldB744TenOrMoreLands is Limited Resources' "as long as ten or more
// lands are on the battlefield".
func S82OldB744TenOrMoreLands(g *game.Game) bool {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsLand() {
			n++
		}
	}
	return n >= 10
}

// S82OldB744LimitedResources is "each player chooses five lands they control
// and sacrifices the rest": each player with more than five lands, in turn
// order, picks the five to keep; then every other land of theirs is sacrificed
// at once (CR 701.21a). A player with five or fewer is not asked.
func S82OldB744LimitedResources(g *game.Game, item *game.StackItem) error {
	var legs []game.ChooseCardsPrompt
	var seats []uuid.UUID
	for _, seat := range ApnapLivingSeats(g) {
		lands := s82OldB744LandsOf(g, seat)
		if len(lands) <= 5 {
			continue
		}
		seats = append(seats, seat)
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Limited Resources — choose five lands you control to keep; the rest are sacrificed",
			Cards:      lands,
			Min:        5,
			Max:        5,
			Zone:       game.ZoneBattlefield,
		})
	}
	if len(legs) == 0 {
		return nil
	}
	return g.ChoosePermanentsInOrderKeyedForEffect(legs, s82OldB744LimitedKeep, item.SourceCardID, game.RunData{Carry: seats})
}

func s82OldB744LandsOf(g *game.Game, seat uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range g.EffectSacrificeCandidatesForEffect(seat, nil) {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			out = append(out, id)
		}
	}
	return out
}

// s82OldB744LimitedKeep is "then sacrifices the rest", keyed. Carry is every
// seat that was asked, in turn order.
var s82OldB744LimitedKeep = game.RegisterPromptRunThen("prompt-run/s82-old-b744-limited-resources-keep", func(g *game.Game, r game.RunDone) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range r.Picks().Cards() {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range r.Carry {
		for _, id := range s82OldB744LandsOf(g, seat) {
			if !kept[id] {
				rest = append(rest, id)
			}
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.SacrificeAllKeyedForEffect(r.Source, rest, game.KeyedTail{})
})

// S82OldB744CrashingBoarsAttack is "whenever this creature attacks, defending
// player chooses an untapped creature they control. That creature blocks this
// creature this turn if able." A choice, not a target: the defending player
// picks among the untapped creatures they control as the trigger resolves, so
// hexproof or protection on a candidate doesn't matter. With none, nothing
// happens.
func S82OldB744CrashingBoarsAttack(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || SourceIsNewObject(g, item) {
		return nil
	}
	defender := B17DefendingPlayer(g, item.Trigger.Event)
	p := g.PlayerByIDForEffect(defender)
	if defender == uuid.Nil || p == nil || p.Eliminated || len(s82OldB744UntappedCreatures(g, defender)) == 0 {
		return nil
	}
	return ChoosePermanents{
		Player:   defender,
		Question: "Crashing Boars — choose an untapped creature you control; it blocks the Boars this turn if able",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return s82OldB744UntappedCreatures(g, of), 1, 1
		},
		Of:    []uuid.UUID{defender},
		Keyed: s82OldB744CrashingBoarsChosen,
	}.Apply(NewContext(g, item))
}

func s82OldB744UntappedCreatures(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && !c.Tapped {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

var s82OldB744CrashingBoarsChosen = RegisterPermanentsPick("prompt-run/s82-old-b744-crashing-boars-block", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		if SourceIsNewObject(ctx.Game, ctx.Item) {
			return nil
		}
		ref, ok := ctx.Game.PermanentRefForEffect(ctx.Item.SourceCardID)
		if !ok {
			return nil
		}
		for _, id := range picked.Cards() {
			if err := (BlocksAttackerUntilEOT{Blocker: id, Attacker: ref, Label: "Crashing Boars — blocks the Boars this turn if able"}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})

// S82OldB744ReturnThisToHand is Jackalope Herd's "return this creature to its
// owner's hand", if it is still the permanent that triggered.
func S82OldB744ReturnThisToHand(g *game.Game, item *game.StackItem) error {
	return ThisStillHere(func(g *game.Game, item *game.StackItem) error {
		return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
	})(g, item)
}
