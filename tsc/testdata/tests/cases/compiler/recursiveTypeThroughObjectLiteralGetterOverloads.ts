// @strict: true
// @noEmit: true

// Whether a get accessor has been resolved records what has been asked for so far, not anything
// about the program, so skipping one on that alone makes an ordinary comparison depend on
// resolution order. This picked the numeric overload for an object whose only member is a string.

declare function pick(x: { [k: string]: number }): "num";
declare function pick(x: object): "obj";

const inferred = pick({
    get s() {
        return "hello";
    },
});
const mustBeObj: "obj" = inferred;

const annotated = pick({
    get s(): string {
        return "hello";
    },
});
const alsoObj: "obj" = annotated;

const numeric = pick({
    get n() {
        return 1;
    },
});
const mustBeNum: "num" = numeric;
