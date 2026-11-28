package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b818_helpers.go — Alliances, +1/+1 and other counters, cumulative
// upkeep (batch 818).

// OldB818SwampMountainBlackOrRedBecameTapped is Royal Decree's "whenever a
// Swamp, Mountain, black permanent, or red permanent becomes tapped": a tap
// event, or an attack that left the attacker tapped, of such a permanent.
// Types and colours are read as the permanent is now (CR 613).
func OldB818SwampMountainBlackOrRedBecameTapped(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventTapCard, game.EventAttack:
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !OnBattlefield(g, ev.CardID) {
		return false
	}
	if ev.Kind == game.EventAttack && !c.Tapped {
		return false
	}
	return c.HasSubtypeWord("Swamp") || c.HasSubtypeWord("Mountain") || c.HasColorWord("B") || c.HasColorWord("R")
}

// OldB818DamageToTheTappedPermanentsController is "this enchantment deals N
// damage to that permanent's controller": the controller the tapped
// permanent has as the trigger resolves.
func OldB818DamageToTheTappedPermanentsController(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
		if !ok {
			return nil
		}
		if p := g.PlayerByIDForEffect(c.Controller); p == nil || p.Eliminated {
			return nil
		}
		ctx := NewContext(g, item)
		return DealDamage{Source: ctx.Source(), Target: c.Controller, Amount: n}.Apply(ctx)
	}
}

// OldB818ThatPlayerExilesTheirLibrary is Thought Lash's "that player exiles
// all cards from their library": the player who did not pay, named by the
// trigger's event.
func OldB818ThatPlayerExilesTheirLibrary(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	player := item.Trigger.Event.Actor
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Library == nil || p.Library.Size() == 0 {
		return nil
	}
	return g.ExileTopKeyedForEffect(player, player, p.Library.Size(), game.KeyedTail{})
}

// OldB818ThreeLifePerVelocityCounter is Tornado's "Pay 3 life for each
// velocity counter on this enchantment", read once at announce.
var OldB818ThreeLifePerVelocityCounter = game.LifeCount("three-per-velocity-counter",
	func(g *game.Game, _, source uuid.UUID) int {
		return 3 * CounterCountOn(g, source, "velocity")
	})

// OldB818VerminInfects is Diseased Vermin's combat damage trigger: an
// infection counter on it, and the player it hit is recorded.
func OldB818VerminInfects(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if item.Trigger != nil && item.Trigger.Event.Target != uuid.Nil {
		g.AddPlayerNoteForEffect(item.SourceCardID, item.Trigger.Event.Target, oldB818DamagedNote)
	}
	return AddCounter{Target: item.SourceCardID, Kind: oldB818InfectionCounter, N: 1}.Apply(ctx)
}

const (
	oldB818DamagedNote      = "dealt damage by it"
	oldB818InfectionCounter = "infection"
)

// OldB818PreviouslyDamagedOpponent is "target opponent previously dealt
// damage by it": an opponent the source has a player note for.
func OldB818PreviouslyDamagedOpponent(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
	noted := source.PlayersNoted(oldB818DamagedNote)
	return TargetPlayer("target opponent previously dealt damage by it",
		func(_ *game.Game, caster uuid.UUID, p *game.Player) bool {
			if p.ID == caster {
				return false
			}
			for _, id := range noted {
				if id == p.ID {
					return true
				}
			}
			return false
		})
}

// OldB818VerminDealsInfectionDamage is the upkeep damage: X is the number of
// infection counters on it as the ability resolves.
func OldB818VerminDealsInfectionDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			target = t.ID
		}
	}
	x := CounterCountOn(g, item.SourceCardID, oldB818InfectionCounter)
	if target == uuid.Nil || x <= 0 {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: target, Amount: x}.Apply(ctx)
}

// PrintedOldB818Splinter is Splintering Wind's 1/1 green Splinter with flying
// and "Cumulative upkeep {G}." and "When it leaves the battlefield, it deals 1
// damage to you and each creature you control."
func PrintedOldB818Splinter() TokenTemplate {
	return TokenTemplate{
		Slug: "splinter",
		Card: game.Card{
			Name:      "Splinter",
			TypeLine:  "Token Creature — Splinter",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"G"},
			Keywords:  []string{"flying"},
		},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Splinter — cumulative upkeep {G}", "{G}"),
			WhenThisLeaves("Splinter — it deals 1 damage to you and each creature you control", OldB818SplinterDamage),
		},
		Text: "Flying\nCumulative upkeep {G}\nWhen this creature leaves the battlefield, it deals 1 damage to you and each creature you control.",
	}
}

// OldB818SplinterDamage is the leaves-the-battlefield damage: "you" is the
// token's controller as it left, "it" the token (last-known information), and
// the damage is dealt at once to you and to each creature you control.
func OldB818SplinterDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := item.Controller
	targets := []uuid.UUID{}
	if p := g.PlayerByIDForEffect(you); p != nil && !p.Eliminated {
		targets = append(targets, you)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() && c.InstanceID != item.SourceCardID {
			targets = append(targets, c.InstanceID)
		}
	}
	if len(targets) == 0 {
		return nil
	}
	return g.DealDamageEachKeyedForEffect(ctx.Source(), targets, 1, game.KeyedTail{})
}
