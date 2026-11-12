package game

// backup.go — which keyword ability a triggered ability on the stack is
// (#5651), so "becomes the target of a backup ability" (Mirror-Shield
// Hoplite) can tell a backup ability from every other trigger.
//
// CR 702.165a: "Backup is a triggered ability." On the stack it is an
// ordinary enters trigger with a target; what makes it a BACKUP ability
// is which catalog row it came from. That row already says so —
// TriggeredAbility.Keyword, stamped by the keyword's constructor
// (effects.Backup) — and the stack item already names its row
// (Params.Ability, ADR 0041 P9). So nothing new is carried on the item:
// the keyword is read through the ref, from the running catalog, and a
// restored item or a copy (CR 707.10, which keeps Body and Params)
// answers exactly as the original did.

// KeywordBackup is the machine-readable name a backup trigger carries in
// TriggeredAbility.Keyword (CR 702.165). It is not a canonicalKeywords
// token: backup has no keyword chip, so the trigger keeps its ability
// row (ability_rows.go).
const KeywordBackup = "backup"

// targeterKeyword is the keyword ability a stack item is, or "" when it
// is not a triggered ability built by a keyword's constructor. Only an
// item that names a catalog row can answer (catalog/triggered); a spell,
// an activated ability, a keyed engine trigger and an unkeyed legacy
// Build all read empty, as does an item whose row the running catalog no
// longer has.
func targeterKeyword(item *StackItem) string {
	if item == nil || item.Kind != StackItemTriggered || item.Body != CatalogTriggeredBodyKey || item.Params.Ability == nil {
		return ""
	}
	row, _, outcome := resolveTriggeredAbilityRef(*item.Params.Ability)
	if outcome == abilityRefLost {
		return ""
	}
	return row.Keyword
}
