package heuristic

// aura_host.go prices an answer to CR 303.4f's aura_host prompt (#3158):
// what an Aura put onto the battlefield without being cast enchants as it
// enters. One candidate per answer, a card or a player.
//
// Three cases, read off where the candidate is:
//
//   - A CARD IN A GRAVEYARD. The only Auras that may enchant one are the
//     reanimation family (Animate Dead, Dance of the Dead, ADR 0154): the
//     Aura's enters trigger puts that card onto the battlefield under the
//     bot's control, whoever owned it. So it is worth what the card is
//     worth on the battlefield — its creature value — and the bot takes
//     the best creature in any graveyard. cardValue's "can't cast it for
//     a while" discount does not apply: nothing is cast.
//   - A PERMANENT OR A PLAYER. The prompt carries no oracle text, so the
//     policy reads the Aura the way it reads an Aura SPELL's target
//     (targetsValue): an opponent's permanent is priced as answered, and
//     the bot's own as the small OwnPermanentTarget. A harmful Aura
//     (Pacifism, a Curse) therefore goes on an opponent's best creature or
//     onto the opponent most worth hurting. Among the bot's own permanents
//     — the only candidates of an "enchant creature you control" Aura —
//     the creature worth most wins the tie.
//
// A buff Aura with an unrestricted clause is priced as a harmful one, the
// same gap the cast path has: nothing on the wire says which an Aura is.

// auraHostValue prices one aura_host answer.
func (st *state) auraHostValue(cfg Config, t *targetRef) (float64, string) {
	if t == nil {
		return 0, "enchant nothing"
	}
	if t.Kind == "card" {
		if c := st.graveyard[t.ID]; c != nil {
			if isCreature(c) {
				return st.w.CreatureValue(c), "reanimate the best creature"
			}
			return st.w.permanentValue(c), "enchant the best card in a graveyard"
		}
		if c := st.bf[t.ID]; c != nil && c.Controller == st.me {
			v := cfg.OwnPermanentTarget
			if isCreature(c) {
				v += 0.001 * st.w.CreatureValue(c)
			}
			return v, "enchant our best permanent"
		}
	}
	return st.targetsValue(cfg, []targetRef{*t}), "enchant the best host"
}
