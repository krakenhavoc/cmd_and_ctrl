package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b024_helpers.go — Secret Lair Drop, pump and combat tricks (S82
// older sets, batch 024, #5462): the shared shapes of Squirrel Wrangler,
// Seven Dwarves, Rashel, Depala and Jurin.

// S82OldB024SquirrelsGetPlusOne is Squirrel Wrangler's second ability: "Squirrel
// creatures get +1/+1 until end of turn". Every Squirrel creature on the
// battlefield as it resolves, whoever controls it (CR 611.2c: the set is fixed
// now, so a Squirrel that arrives later gets nothing).
func S82OldB024SquirrelsGetPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	squirrel := ctx.TextSubtype("Squirrel") // a word of the source's text (CR 612.2)
	return BoostUntilEOT{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.IsCreature() && c.HasSubtype(squirrel)
		},
		Power: 1, Toughness: 1,
		Label: "Squirrel Wrangler — Squirrel creatures get +1/+1",
	}.Apply(ctx)
}

// S82OldB024PlusOnePerOtherNamedYouControl is Seven Dwarves' "This creature
// gets +1/+1 for each other creature named <name> you control", a layer-7c
// bonus counted live. The name is the printed one, read as it is now on each
// creature (game.HasName), so a copy of it counts and a Dwarf with another name
// does not.
func S82OldB024PlusOnePerOtherNamedYouControl(name string) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Gets +1/+1 for each other creature named " + name + " you control.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			for _, o := range g.BattlefieldCardsForEffect() {
				if o.InstanceID != source.InstanceID && o.Controller == source.Controller && o.IsCreature() && game.HasName(o, name) {
					n++
				}
			}
			c.Power += n
			c.Toughness += n
		},
	}
}

// S82OldB024AurasYouControl is the subject of "Auras you control have …"
// (Rashel): every Aura permanent its controller has, read through the layers.
func S82OldB024AurasYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasSubtypeWord("Aura")
}

// S82OldB024VehiclesYouControlPlusOne is Depala's "Each Vehicle you control
// gets +1/+1 as long as it's a creature": a layer-7c bonus for each Vehicle its
// controller has that is a creature right now (CR 301.7).
func S82OldB024VehiclesYouControlPlusOne() game.StaticAbility {
	return game.StaticAbility{
		Label:    "Each Vehicle you control gets +1/+1 as long as it's a creature.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.Controller == source.Controller && target.IsCreature() && target.HasSubtypeWord("Vehicle")
		},
		Apply: PumpBy(1, 1),
	}
}

// S82OldB024DepalaAsk is Depala's trigger body: "you may pay {X}". X is chosen
// as it resolves and paid like any may-pay, then the paid branch reveals.
func S82OldB024DepalaAsk(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:  "{X}",
		Label: "Depala, Pilot Exemplar",
		Buys:  "reveal the top X cards of your library; put all Dwarf and Vehicle cards among them into your hand",
		Unit:  game.PayAmountCards,
		Then:  S82OldB024DepalaPaid,
	}.Apply(NewContext(g, item))
}

// S82OldB024DepalaPaid is the paid half: reveal the top X cards of the library,
// put all Dwarf and Vehicle cards revealed this way into your hand, then the
// rest on the bottom in a random order. No prompt (there is no choice), and the
// rest is a registered take so a table waiting on it is a restore point.
var S82OldB024DepalaPaid = PayUnlessBranches("pay-unless/s82-old-b024-depala-reveal",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		player := ctx.Controller()
		const label = "Depala, Pilot Exemplar — reveal the top X cards"
		revealed := ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), r.Number, label)
		dwarf, vehicle := ctx.TextSubtype("Dwarf"), ctx.TextSubtype("Vehicle")
		return TakeFromLibraryToHand{
			Player: player,
			Cards:  revealed,
			Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.HasSubtype(dwarf) || c.HasSubtype(vehicle)
			},
			All:   true,
			Label: label,
			Key:   TakeRestBottomRandom,
		}.applyKeyed(ctx, player, ctx.Source())
	}, nil)

// S82OldB024JurinPump is Jurin's attack trigger: "each creature you control
// attacking a player gets +1/+0 until end of turn for each creature that player
// controls". Each attacker is judged against the player IT attacks (an attacker
// at a planeswalker or battle is not attacking a player), counted as the
// trigger resolves, and the bonus is fixed from then on.
func S82OldB024JurinPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	type pump struct {
		id uuid.UUID
		n  int
	}
	var pumps []pump
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() || c.AttackingTarget == uuid.Nil {
			continue
		}
		defender := c.AttackingTarget
		if p := g.PlayerByIDForEffect(defender); p == nil || p.Eliminated {
			continue
		}
		n := 0
		for _, o := range g.BattlefieldCardsForEffect() {
			if o.Controller == defender && o.IsCreature() {
				n++
			}
		}
		if n > 0 {
			pumps = append(pumps, pump{c.InstanceID, n})
		}
	}
	for _, p := range pumps {
		if err := (BoostUntilEOT{Target: p.id, Power: p.n, Label: "Jurin, Leading the Charge — +1/+0 for each creature the defending player controls"}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
