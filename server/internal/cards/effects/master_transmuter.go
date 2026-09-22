package effects

// Master Transmuter — Artifact Creature — Human Artificer {3}{U},
// 1/2:
//
//	"{U}, {T}, Return an artifact you control to its owner's hand:
//	 You may put an artifact card from your hand onto the
//	 battlefield."
//
// The card #1213's seam row names, and the one that exercises the
// awkward corner of the component: Master Transmuter IS an artifact
// you control, so she is a legal pick for her own cost. Returning
// herself taps her, bounces her and still resolves the ability — a
// real line, because the artifact that comes down can be worth more
// than she is. The activation path drops its handle on the source
// after the cost is paid for exactly this reason, the same way it
// does after a sacrifice.
//
// Three things the cost gets right by being a cost:
//
//   - CR 118.3 — with no artifact on the battlefield she cannot
//     activate, even with {U} up and untapped. (She is always one
//     herself, so in practice the clause is only unpayable while she
//     is not on the battlefield.)
//   - CR 601.2h / 602.2b — the bounce is paid in one indivisible
//     step at announce, so the artifact is in hand and the ability is
//     on the stack before anyone gets priority. A "whenever an
//     artifact leaves the battlefield" watcher triggers and resolves
//     ABOVE it.
//   - the put is an EFFECT, so it happens on resolution and can be
//     answered. The artifact the cost returned is a legal thing to
//     put back down with it, which is the blink-flavoured line the
//     card is famous for.
//
// The put itself is "you may", so declining is a real answer, and it
// puts the card onto the battlefield without casting it — no mana
// cost, no cast triggers, and "when this enters" still fires.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da46786e-28df-4638-ab3d-121011d2f150",
		Name:         "Master Transmuter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{U}, {T}, Return an artifact you control to its owner's hand: You may put an artifact card from your hand onto the battlefield.",
			Cost: Plus(
				ManaCost("{U}"),
				TapCost(),
				ReturnAPermanentToHand("an artifact you control", Artifact()),
			),
			Effect: Do(PutFromHandOntoBattlefield{
				Match:    Artifact(),
				Optional: true,
				Label:    "Master Transmuter — you may put an artifact card from your hand onto the battlefield",
			}),
		}},
	})
}
