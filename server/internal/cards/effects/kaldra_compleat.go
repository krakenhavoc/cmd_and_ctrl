package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kaldra Compleat — Legendary Artifact — Equipment for {7}:
//
//	"Living weapon
//	 Indestructible
//	 Equipped creature gets +5/+5 and has first strike, trample,
//	 indestructible, haste, and 'Whenever this creature deals combat
//	 damage to a creature, exile that creature.'
//	 Equip {7}"
//
// Seven mana for an indestructible hasty 5/5 first striker that
// attacks the turn it lands, and re-arms itself every time the Germ
// is answered, because the Equipment's own indestructible survives
// anything that kills the body.
//
// THE LAST CLAUSE IS A GRANTED TRIGGER (ADR 0093). "Whenever this
// creature deals combat damage to a creature, exile that creature" is
// a quoted ability the Equipment gives its host, so it is a catalog
// bundle named by a layer-6 grant (GrantAbilitiesToAttached), exactly
// as Thornbite Staff's untap trigger is. "This creature" is the HOST:
// the trigger watches the host's combat damage, goes on the stack
// under the host's controller, and a removal on the host takes it.
// Moving Kaldra moves it.
//
// "That creature" is the object that was dealt the damage. Its epoch
// is read as the trigger is built (the ADR 0041 P9 fill-in Build, as
// Ares, God of War reads its dead creature's), so a creature that
// died to the damage and came back before the trigger resolves is a
// new object (CR 400.7) and is not exiled. A creature the damage did
// not kill — indestructible, or big enough — is.
//
// The other four grants are real and all four have live consumers:
// first strike splits the damage step, trample assigns the excess,
// indestructible is read by the destruction path and the damage
// state-based action, haste clears the summoning-sickness gate on a
// Germ created this turn — which is what makes living weapon plus
// haste a seven-mana attack rather than a seven-mana investment.
//
// The Equipment's OWN indestructible is a printed keyword rather
// than a static, the Darksteel Plate distinction: true of the
// Equipment in every zone, and it survives the Germ dying.
func init() {
	Register(Spec{
		OracleID:        "7359e82b-db79-488d-a1d4-75a00f12a4cf",
		Name:            "Kaldra Compleat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Grants: []AbilityGrant{{
			Key: kaldraCompleatExile,
			Triggered: []game.TriggeredAbility{{
				Watches:   []game.EventKind{game.EventDealDamage},
				AppliesTo: thisDealtCombatDamageToACreature,
				Key:       kaldraCompleatExileLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, kaldraCompleatExileLabel)
					item.Params.Object = damagedCreatureRef(g, ev.Target)
					return item
				},
				Effect: exileDamagedCreature,
			}},
			Text: "Whenever this creature deals combat damage to a creature, exile that creature.",
		}},
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Kaldra Compleat"),
		},
		Static: []game.StaticAbility{
			PumpAttached(5, 5),
			GrantToAttached("first strike", "trample", "indestructible", "haste"),
			GrantAbilitiesToAttached(kaldraCompleatExile),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{7}"),
		},
	})
}

// kaldraCompleatExile is the granted bundle's key, and
// kaldraCompleatExileLabel the trigger's stack label.
const (
	kaldraCompleatExile      = "kaldra-compleat/exile"
	kaldraCompleatExileLabel = "Kaldra Compleat — exile the creature this creature dealt combat damage to"
)
