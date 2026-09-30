package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cityscape Leveler — Artifact Creature — Construct {8}, 8/8:
//
//	"Trample
//	 When you cast this spell and whenever this creature attacks,
//	 destroy up to one target nonland permanent. Its controller
//	 creates a tapped Powerstone token.
//	 Unearth {8}"
//
// One printed ability with two trigger conditions — a cast (while the
// Leveler is still a spell on the stack, FromStack's narrow scan) and
// an attack (once it is a permanent) — so it is one TriggeredAbility
// watching both event kinds rather than two declarations, the same
// shape Sun Titan's "enters or attacks" takes. Build fills in the
// controller only for the cast leg, because that leg's source is a
// spell whose Controller field is not reliably the caster
// (Desolation Twin, Oblivion Sower); the attack leg's source is
// already a permanent with the right Controller.
//
// "Up to one" needs no prompt when nothing is chosen (CR 603.3d), and
// the destroyed permanent's controller is read BEFORE the destroy —
// after that it is a card in a graveyard with no controller to hand a
// token to.
//
// Unearth is the constructor: the return, the haste, the end-step
// exile and the leaves-the-battlefield redirect all come with the
// keyword.
//
// Inherited gap, not this card's own: the Powerstone token this
// creates can tap for mana to cast any spell — its printed
// restriction to artifact spells isn't modelled (see PowerstoneToken
// in tokens.go). Declared here rather than fixed, since the
// restriction lives on the shared token template, not on this card.
func init() {
	Register(Spec{
		OracleID:     "d4d65797-2b92-4265-9169-133120c86c7f",
		Name:         "Cityscape Leveler",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Powerstone token this creates can pay for any spell — its printed restriction to artifact spells isn't implemented.",
		},
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast, game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				controller := source.Controller
				if ev.Kind == game.EventCast {
					controller = ev.Actor
				}
				return &game.StackItem{
					Kind:         game.StackItemTriggered,
					Controller:   controller,
					Owner:        controller,
					SourceCardID: source.InstanceID,
					Label:        "Cityscape Leveler — destroy up to one target nonland permanent",
				}
			},
			Effect: cityscapeLevelerDestroyAndPowerstone,
		}},
		Activated: []ActivatedAbility{Unearth("{8}")},
	})
}

// cityscapeLevelerDestroyAndPowerstone reads the target's controller
// before destroying it — a graveyard card has none to hand a
// Powerstone to — then destroys it and hands that controller the
// token. "Up to one" left unfilled is a no-op (CR 603.3d already kept
// this off the stack in that case; this is the belt for it).
func cityscapeLevelerDestroyAndPowerstone(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	target, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	controller := target.Controller
	if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return CreateTokenAdvanced{
		Controller: controller,
		Spec:       Token(PowerstoneToken()).EntersTapped(),
		N:          1,
	}.Apply(ctx)
}
