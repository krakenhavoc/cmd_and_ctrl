package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foot Chopper — {1}{B} Artifact — Equipment (Teenage Mutant Ninja
// Turtles Commander, #5011):
//
//	"When this Equipment enters, create a 1/1 black Ninja creature token,
//	 then attach this Equipment to it.
//	 Equipped creature has flying.
//	 Whenever equipped creature deals combat damage to a player, you may
//	 sacrifice it. If you do, draw cards equal to its power.
//	 Equip {2}"
//
// The entry is Dwarven Shortsword's token-then-attach; the flying grant is
// Cursed Windbreaker's. The combat-damage trigger remembers the creature
// that dealt the damage as an object (CR 400.7): "it" is that creature
// even if Foot Chopper has moved since, and a creature that left (or came
// back as a new object) can't be sacrificed. You can sacrifice only a
// creature you control (CR 701.21a). The draw counts the creature's power
// as it last existed on the battlefield (the printed ruling, CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1eed017a-342a-47a8-9395-890ef0bd1eb2",
		Name:         "Foot Chopper",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Foot Chopper — create a 1/1 black Ninja creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "1/1 black Ninja", true)
				}),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				},
				Key: footChopperLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, footChopperLabel)
					epoch := -1
					if c, ok := g.LookupCardForEffect(ev.Source); ok {
						epoch = c.ObjectEpoch
					}
					item.Params.Object = game.ObjectRef{ID: ev.Source, Epoch: epoch}
					return item
				},
				Effect: footChopperMaySacrifice,
			},
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

const footChopperLabel = "Foot Chopper — you may sacrifice that creature; if you do, draw cards equal to its power"

// footChopperMaySacrifice asks only while the creature is still the
// object that dealt the damage and is yours to sacrifice.
func footChopperMaySacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref := item.Params.Object
	c, ok := g.LookupCardForEffect(ref.ID)
	if !ok || !onBattlefield(g, ref.ID) || c.ObjectEpoch != ref.Epoch || c.Controller != item.Controller {
		return nil
	}
	return MayChoice{
		Question: "Foot Chopper — sacrifice " + c.Name + " to draw cards equal to its power?",
		YesLabel: "Sacrifice",
		NoLabel:  "Decline",
		Then:     footChopperAnswer,
		Objects:  []game.ObjectRef{ref},
	}.Apply(ctx)
}

// footChopperDraw is "if you do, draw cards equal to its power": told the
// creature that was sacrificed, it reads that creature's power as it last
// existed on the battlefield (CR 608.2h). An on-disk key.
var footChopperDraw = CardTail("foot-chopper/draw-cards-equal-to-sacrificed-power", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	if n := departedCreaturePower(ctx.Game, o.IDs[0]); n > 0 {
		return DrawCards{N: n}.Apply(ctx)
	}
	return nil
})

// footChopperAnswer sacrifices the creature the prompt carries, if it is
// still that object and still yours. An on-disk key.
var footChopperAnswer = MayYes("confirm/foot-chopper-sacrifice", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Objects) == 0 {
		return nil
	}
	ref := r.Objects[0]
	c, ok := ctx.Game.LookupCardForEffect(ref.ID)
	if !ok || !onBattlefield(ctx.Game, ref.ID) || c.ObjectEpoch != ref.Epoch || c.Controller != ctx.Controller() {
		return nil
	}
	return SacrificePermanent{Target: ref.ID, ThenTail: ctx.ItemTail(footChopperDraw)}.Apply(ctx)
})
