// @strict: true
// @noEmit: true

// KNOWN GAP, pinned so it stays visible. This is the single-file case in
// recursiveTypeThroughObjectLiteralGetterPostponedConstraint.ts split across two files: an array of
// schemas is not a Schema, so `bad` violates the constraint on S. Single-file that is reported, via
// the postponed check. Here it is not: the constraint is answered while the getter still stands on
// an absorbed circularity, and by the time the getter has a type the verdict has been reached.
//
// `main` reports two implicit-any circularity errors here instead, which this change removes by
// design. So the net effect on this shape is that invalid code type-checks clean. Fixing it needs the
// verdict itself to be retractable rather than merely uncached, which is a larger change.

// @filename: schema.ts
export interface Schema<O> {
    readonly out: O;
}
export type Shape = Record<string, Schema<any>>;
export declare function object<S extends Shape>(shape: S): Schema<{ [K in keyof S]: S[K]["out"] }> & { shape: S };

export const tree = object({
    get bad() {
        return [tree];
    },
});

// @filename: consumer.ts
import { tree } from "./schema.js";

export const out = tree.out;
