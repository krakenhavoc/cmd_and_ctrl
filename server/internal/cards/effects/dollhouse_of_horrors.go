package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dollhouse of Horrors — Artifact {5} (Innistrad: Crimson Vow):
//
//	"{1}, {T}, Exile a creature card from your graveyard: Create a token
//	 that's a copy of the exiled card, except it's a 0/0 Construct artifact
//	 in addition to its other types and it has "This token gets +1/+1 for
//	 each Construct you control." It gains haste until end of turn.
//	 Activate only as a sorcery."
//
// The exiled card is the cost's (ctx.Exiled, CR 602.2b), copied where it now
// is. The exceptions are part of the copiable values (CR 707.9b): 0/0, the
// Artifact type and Construct subtype, and the quoted ability, a bundle with
// one layer 7c static that counts the Constructs its controller controls
// (the token itself included). "It gains haste until end of turn" is not
// copiable: it is a layer 6 effect pinned to that token for the turn
// (Mishra, Eminent One's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73970475-f899-4e93-8498-286240181295",
		Name:         "Dollhouse of Horrors",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: dollhouseOfHorrorsPumpGrant,
			Static: []game.StaticAbility{{
				Label:    "This token gets +1/+1 for each Construct you control.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := len(permanentsControlledByMatching(g, source.Controller, Subtype("Construct")))
					c.Power += n
					c.Toughness += n
				},
			}},
			Text: "This token gets +1/+1 for each Construct you control.",
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}, Exile a creature card from your graveyard: Create a token that's a copy of the exiled card, except it's a 0/0 Construct artifact in addition to its other types and it has \"This token gets +1/+1 for each Construct you control.\" It gains haste until end of turn. Activate only as a sorcery.",
			Cost: Plus(ManaCost("{1}"), TapCost(),
				ExileFromGraveyard(1, "a creature card", isCreatureCard)),
			SorcerySpeed: true,
			Effect:       dollhouseOfHorrorsCopy,
		}},
	})
}

const dollhouseOfHorrorsPumpGrant = "dollhouse-of-horrors/construct-pump"

func dollhouseOfHorrorsCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	exiled := ctx.Exiled()
	if len(exiled) == 0 {
		return nil
	}
	cursor := b25LastEventSeq(g)
	if err := (CreateTokenCopy{
		Controller: item.Controller,
		Copy:       exiled[0],
		N:          1,
		Except: func(t *game.Card) {
			v := game.PrintedValues{TypeLine: t.TypeLine, GrantedAbilities: t.GrantedAbilities}
			v.AddCardType("Artifact")
			v.AddSubtype("Construct")
			v.GrantAbility(dollhouseOfHorrorsPumpGrant)
			t.TypeLine = v.TypeLine
			t.GrantedAbilities = v.GrantedAbilities
			t.Power, t.Toughness = 0, 0
			t.VariableToughness = false
			t.PrintedPTKnown = true
		},
	}).Apply(ctx); err != nil {
		return err
	}
	for _, token := range b27TokensCreatedByAfter(g, item.Controller, cursor) {
		if err := (ScopedEffectFor{
			Target:   token,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Dollhouse of Horrors — the token gains haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
