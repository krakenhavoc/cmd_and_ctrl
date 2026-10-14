package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emrakul, the Exigent Doom — Legendary Creature — Eldrazi {10}, 12/12:
//
//	"When you cast this spell, untap all lands you control.
//	 Flying, trample
//	 Ward—Sacrifice three permanents.
//	 {3}, Exile this card from your hand: Target land gains "{T}: Add
//	 {C}{C}" until this card is cast from exile. You may cast this card
//	 for as long as it remains exiled."
//
// The cast trigger is WhenYouCastThisSpell (it fires from the stack and
// resolves above the spell, so a countered Emrakul still untaps), and
// untaps every land the CASTER controls as it resolves. Ward is the
// shared ward trigger with a sacrifice-three cost; the three may be any
// permanents, lands included.
//
// # The hand ability (#3520)
//
// "{3}, Exile this card from your hand" is an ordinary activated ability
// that functions from the hand: the exile is a cost paid on activation
// (ExileThis from ZoneHand, which ExileSelfZoneSupported allows since this
// card), so Emrakul is in exile before anyone can respond. It targets a
// land, and on resolution does two things:
//
//   - The land gains "{T}: Add {C}{C}" until THIS CARD is cast from exile.
//     The grant is the ADR 0093 duration grant (GrantAbilitiesFor) with the
//     UntilCardCast duration (game.UntilCardCast, ADR 0063's 2026-10-14
//     amendment): it names the exile OBJECT and ends when CastSpell casts
//     it, after the mana has paid for the cast (ruling 2026-08-21: the land
//     may tap for it while Emrakul is being cast from exile). Emrakul
//     leaving exile uncast does not end it (the other 2026-08-21 ruling),
//     and if Emrakul has already left exile when the ability resolves the
//     grant has nothing to wait for and simply stays (CR 611.2b).
//   - "You may cast this card for as long as it remains exiled" is a stored
//     cast permission over the exiled card with the WhileInZone duration
//     (the airbend shape). It prices nothing: the cast pays {10}, with
//     timing as a creature spell, and it is gone the moment the card leaves
//     exile (CR 400.7).
//
// If the land is an illegal target on resolution the whole ability does
// nothing (CR 608.2b), the exiled card included: it stays exiled and is not
// castable, which is the printed risk of the ability.
//
// No simplification.
const emrakulExigentTapGrant = "emrakul-the-exigent-doom/tap-cc"

func init() {
	Register(Spec{
		OracleID:        "4421ab7d-6d9b-4edd-b5a0-53a8ed84da6f",
		Name:            "Emrakul, the Exigent Doom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Grants: []AbilityGrant{
			TapForManaGrant(emrakulExigentTapGrant, "{C}{C}", "Add {C}{C}", "{T}: Add {C}{C}."),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, Exile this card from your hand: Target land gains \"{T}: Add {C}{C}\" until this card is cast from exile. You may cast this card for as long as it remains exiled.",
			Cost:    Plus(ManaCost("{3}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetPermanent("target land", Land()),
			Effect:  emrakulExigentExile,
		}},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Emrakul, the Exigent Doom — untap all lands you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var lands []game.Card
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsLand() && c.Tapped {
							lands = append(lands, c)
						}
					}
					for _, c := range lands {
						if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
			Ward(WardSacrificeN(3, "three permanents", Permanent()), "Emrakul, the Exigent Doom — ward, sacrifice three permanents"),
		},
	})
}

// emrakulExigentExile is the hand ability's effect: the land gains the
// mana ability until the exiled card is cast, and the controller may cast
// the card for as long as it stays in exile.
func emrakulExigentExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	land, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	card := item.SourceCardID
	d, inExile := g.UntilCardCastDuration(card)
	if !inExile {
		d = game.IndefiniteDuration()
	}
	if err := (GrantAbilitiesFor{
		Target:   land,
		Keys:     []string{emrakulExigentTapGrant},
		Duration: d,
		Label:    "Emrakul, the Exigent Doom — the land taps for {C}{C} until Emrakul is cast from exile",
	}).Apply(ctx); err != nil {
		return err
	}
	if inExile {
		g.GrantCastPermissionOverCardForEffect(card, game.CastPermission{
			Player:     item.Controller,
			Duration:   game.WhileInZoneDuration(),
			CastOnly:   true,
			Source:     card,
			SourceName: "Emrakul, the Exigent Doom",
			Label:      "Emrakul, the Exigent Doom — cast it from exile",
		})
	}
	return nil
}
