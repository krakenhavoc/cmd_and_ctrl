package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Irencrag — Legendary Artifact {2}:
//
//	"{T}: Add {C}.
//	 Whenever a legendary creature you control enters, you may have The
//	 Irencrag become a legendary Equipment artifact named Everflame,
//	 Heroes' Legacy. If you do, it gains equip {3} and 'Equipped creature
//	 gets +3/+3' and loses all other abilities."
//
// #2562 (ADR 0093 amendment 2026-10-10). The change is ONE effect with
// no duration (CR 611.2a; the 2023-09-01 ruling: "The Irencrag's last
// ability lasts indefinitely"), registered as one ScopedEffect record
// pinned to this object, so it ends only if the permanent leaves and
// comes back as a new object (CR 400.7). Its mods, each in its layer, at
// one timestamp:
//
//   - layer 3: named Everflame, Heroes' Legacy (CR 612.8). The legend
//     rule reads the effective name, so a second Irencrag that has not
//     changed is not a legend-rule pair with Everflame, and two
//     Everflames are.
//   - layer 4: a legendary Equipment artifact (CR 205.1a: a set, not an
//     add), so the Equipment's attach checks (CR 301.5c) see an
//     Equipment.
//   - layer 6: loses all its abilities, then gains the two bundles, the
//     removal first so the grants survive it (ADR 0046 §2). Equip is an
//     activated ability of the permanent (CR 702.6a), so the granted row
//     is the host's own: its controller activates it, at sorcery speed,
//     and it attaches the host. The "+3/+3" is a granted layer-7c static
//     (game/granted_statics.go), applied to whatever the host is attached
//     to.
//
// Neither the name nor the grants are copiable (CR 707.2): a copy of
// Everflame is a copy of The Irencrag.
//
// No simplification.
const (
	irencragOracleID      = "8c89ba89-9bb3-4609-873c-6f5a961733ce"
	irencragEverflameName = "Everflame, Heroes' Legacy"
	irencragEquipGrant    = "the-irencrag/everflame-equip"
	irencragPumpGrant     = "the-irencrag/everflame-pump"
)

func init() {
	pump := PumpAttached(3, 3)
	pump.Label = "Equipped creature gets +3/+3."
	Register(Spec{
		OracleID:     irencragOracleID,
		Name:         "The Irencrag",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Grants: []AbilityGrant{
			{Key: irencragEquipGrant, Activated: []ActivatedAbility{EquipAbility("{3}")}, Text: "Equip {3}"},
			{Key: irencragPumpGrant, Static: []game.StaticAbility{pump}, Text: "Equipped creature gets +3/+3."},
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.IsLegendary()
			},
			Key:            "The Irencrag — become Everflame, Heroes' Legacy",
			Effect:         irencragBecomeEverflame,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Have The Irencrag become Everflame, Heroes' Legacy?"},
		}},
	})
}

// irencragBecomeEverflame is the trigger's resolution: the indefinite
// change, pinned to this object. A source that has left, or come back as
// a new object, changes nothing (CR 400.7, 608.2b).
func irencragBecomeEverflame(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	if !onBattlefield(g, self) || sourceIsNewObject(g, item) {
		return nil
	}
	return ScopedEffectFor{
		Target: self,
		Mods: []game.Mod{
			game.SetNameMod(irencragEverflameName),
			game.SetTypesMod([]string{"Artifact"}, []string{"Equipment"}, "Legendary"),
			game.LoseAllAbilitiesMod(),
			game.GrantAbilitiesMod(irencragEquipGrant, irencragPumpGrant),
		},
		Duration: g.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "The Irencrag — Everflame, Heroes' Legacy",
	}.Apply(NewContext(g, item))
}
