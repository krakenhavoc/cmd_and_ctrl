package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serra Paragon — Creature — Angel {2}{W}{W}, 3/4 (Secrets of Strixhaven
// Commander, #4834):
//
//	"Flying
//	 Once during each of your turns, you may play a land from your
//	 graveyard or cast a permanent spell with mana value 3 or less from
//	 your graveyard. If you do, it gains "When this permanent is put into a
//	 graveyard from the battlefield, exile it and you gain 2 life.""
//
// Flying is the printed keyword. The second ability is a STANDING graveyard
// permission (ADR 0066, Muldrotha's and Gisa and Geralf's shape):
//
//   - Filter.PermanentOnly: a land card (played, CR 305.1) or a permanent
//     card, but never an instant or sorcery. Mana value 3 or less is
//     MaxSpellManaValue, judged on the face actually cast; a land has mana
//     value 0.
//   - ONE use per turn however it is spent: PerType holds the single
//     pseudo-type PermissionTypePermanent, so playing a land spends the same
//     use as casting a spell. It is kept per Serra Paragon OBJECT, so a
//     second one that arrives later in the turn opens another use, and
//     Timing is TimingYourTurnOnly, so each card keeps its own timing and a
//     land still needs a main phase and an empty stack. A play also spends
//     the turn's land drop (CR 305.2).
//   - The rider is the permission's GainsAbilities: a permanent spell cast
//     through it, or a land played through it, gains the bundle below for
//     as long as it is on the battlefield as that object (CR 400.7a,
//     CR 611.2a). A land entering after a pause on its own replacement (a
//     shockland's life payment) still gains it, as the entry finishes.
//     A card whose own text opens the graveyard (flashback, a Gravecrawler)
//     is cast by that text, uses nothing here and gains nothing.
//
// The granted "When this permanent is put into a graveyard from the
// battlefield, exile it and you gain 2 life." is the permanent's own trigger
// (ADR 0093): it fires on any trip from the battlefield to a graveyard, is
// controlled by the permanent's controller, exiles the card only while it
// is still the object that went to the graveyard (CR 400.7), and gains the
// life regardless, since the two are separate instructions.
//
// No simplification.
const serraParagonGrant = "serra-paragon/exile-and-gain-life"

func init() {
	three := 3
	Register(Spec{
		OracleID:        "ecd8d7f4-7b11-4097-afc0-ef3c81edd90c",
		Name:            "Serra Paragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CastPermissions: []game.CastPermission{{
			Zone:              game.ZoneGraveyard,
			Filter:            game.PermissionFilter{PermanentOnly: true},
			MaxSpellManaValue: &three,
			PerType:           []string{game.PermissionTypePermanent},
			Timing:            game.TimingYourTurnOnly,
			GainsAbilities:    []string{serraParagonGrant},
			Label:             "Once during each of your turns, play a land from your graveyard or cast a permanent spell with mana value 3 or less from your graveyard (Serra Paragon)",
		}},
		Grants: []AbilityGrant{{
			Key: serraParagonGrant,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Serra Paragon — this permanent was put into a graveyard from the battlefield: exile it and you gain 2 life",
					serraParagonExileAndGain),
			},
			Text: "When this permanent is put into a graveyard from the battlefield, exile it and you gain 2 life.",
		}},
	})
}

// serraParagonExileAndGain is the granted trigger's body: the card that
// went to the graveyard, while it is still that object there, is exiled;
// its controller gains 2 life either way.
func serraParagonExileAndGain(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	dead := item.Trigger.Event.CardID
	if diedCardStillInGraveyard(ctx, dead) {
		if err := (ExileTarget{Target: dead}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
}
