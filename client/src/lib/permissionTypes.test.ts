import { describe, expect, it } from "vitest";
import {
  needsPermissionTypePicker,
  permissionTypeLabel,
  permissionTypeVerb,
  permissionTypesOf,
} from "./permissionTypes";
import { cardAsFace } from "./faces";
import { applyCastChoices } from "./targeting";
import type { CardView } from "./protocol";

// permissionTypes.test.ts — #2167: Muldrotha, the Gravetide's "If a card
// has multiple permanent types, choose one as you play it". The server
// lists the types a face may spend; the cast chain asks only when there
// are two or more, and sends the answer as `permission_type`.

function card(over: Partial<CardView> = {}): CardView {
  return {
    instance_id: "card-1",
    name: "Ornithopter",
    owner: "me",
    controller: "me",
    ...over,
  } as CardView;
}

describe("permission types", () => {
  it("asks only when the face has two or more types left", () => {
    expect(needsPermissionTypePicker(card())).toBe(false);
    expect(needsPermissionTypePicker(card({ permission_types: ["creature"] }))).toBe(false);
    expect(needsPermissionTypePicker(card({ permission_types: ["artifact", "creature"] }))).toBe(
      true,
    );
  });

  it("keeps the server's ranked order", () => {
    expect(permissionTypesOf(card({ permission_types: ["creature", "artifact"] }))).toEqual([
      "creature",
      "artifact",
    ]);
  });

  it("labels the rows and the confirm", () => {
    const c = card();
    expect(permissionTypeLabel("artifact")).toBe("Artifact");
    expect(permissionTypeVerb(c, "artifact")).toBe("Cast Ornithopter as an artifact");
    expect(permissionTypeVerb(c, "creature")).toBe("Cast Ornithopter as a creature");
    expect(permissionTypeVerb(c, "land")).toBe("Play Ornithopter as a land");
  });

  it("sends the chosen type as permission_type, and nothing when none was asked", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { fromZone: "graveyard", permissionType: "artifact" });
    expect(params.permission_type).toBe("artifact");
    const bare: Record<string, unknown> = {};
    applyCastChoices(bare, { fromZone: "graveyard" });
    expect("permission_type" in bare).toBe(false);
  });

  it("swaps the list with the face, so each half asks about itself", () => {
    const mdfc = card({
      permission_types: ["creature"],
      faces: [
        { name: "Front", type_line: "Creature", permission_types: ["creature"] },
        {
          name: "Back",
          type_line: "Artifact Creature",
          permission_types: ["artifact", "creature"],
        },
      ],
    } as Partial<CardView>);
    expect(needsPermissionTypePicker(cardAsFace(mdfc, 0))).toBe(false);
    expect(needsPermissionTypePicker(cardAsFace(mdfc, 1))).toBe(true);
  });
});
