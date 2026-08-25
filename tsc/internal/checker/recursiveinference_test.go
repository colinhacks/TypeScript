package checker_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

func checkSourceText(t *testing.T, content string) []string {
	t.Helper()
	fs := vfstest.FromMap(map[string]string{
		"/main.ts": content,
		"/tsconfig.json": `{
			"compilerOptions": { "strict": true, "noEmit": true },
			"files": ["main.ts"]
		}`,
	}, false /*useCaseSensitiveFileNames*/)
	fs = bundled.WrapFS(fs)

	host := compiler.NewCompilerHost("/", fs, bundled.LibPath(), nil, nil, nil)
	parsed, errors := tsoptions.GetParsedCommandLineOfConfigFile("/tsconfig.json", &core.CompilerOptions{}, nil, host, nil)
	assert.Equal(t, len(errors), 0, "Expected no errors in parsed command line")

	p := compiler.NewProgram(compiler.ProgramOptions{Config: parsed, Host: host})
	p.BindSourceFiles()
	file := p.GetSourceFile("/main.ts")
	var messages []string
	for _, d := range p.GetSemanticDiagnostics(t.Context(), file) {
		messages = append(messages, fmt.Sprintf("TS%d at %d: %s", d.Code(), d.Pos(), string(d.MessageKey())))
	}
	return messages
}

// Skipping a get accessor that has no type yet is only correct where it genuinely cannot be worked
// out. Whether an accessor has been resolved records what has been asked for so far, nothing about
// the program, so skipping on that alone makes an ordinary comparison depend on resolution order.
// This selected the numeric overload for an object whose only member is a string.
func TestUnresolvedAccessorDoesNotChangeOverloadResolution(t *testing.T) {
	t.Parallel()

	messages := checkSourceText(t, `
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
`)
	assert.Equal(t, len(messages), 0, "expected no errors, got: "+strings.Join(messages, "; "))
}

// A getter in an object literal argument may name the very variable being declared. Inferring the
// type argument must not force that getter, which would report a circularity and fix it at `any`.
// The recursive type has to come out whole, and the constraint has to stay enforced.
func TestRecursiveObjectLiteralGetterInference(t *testing.T) {
	t.Parallel()

	messages := checkSourceText(t, `
interface Internals<out O = unknown> {
	readonly out: O;
}
interface Schema {
	readonly internals: Internals;
}
type Out<T extends Schema> = T["internals"]["out"];

interface ArrayInternals<T extends Schema> extends Internals<Out<T>[]> {}
interface ArraySchema<T extends Schema = Schema> extends Schema {
	readonly internals: ArrayInternals<T>;
}
interface StringInternals extends Internals<string> {}
interface StringSchema extends Schema {
	readonly internals: StringInternals;
}

type Shape = Readonly<{ [k: string]: Schema }>;
interface ObjectInternals<S extends Shape> extends Internals<{ [K in keyof S]: Out<S[K]> }> {}
interface ObjectSchema<S extends Shape = Shape> extends Schema {
	readonly internals: ObjectInternals<S>;
}

declare function object<S extends Shape>(shape: S): ObjectSchema<S>;
declare function array<T extends Schema>(element: T): ArraySchema<T>;
declare function text(): StringSchema;

const node = object({
	name: text(),
	get children() {
		return array(node);
	},
});

type TreeNode = Out<typeof node>;
declare const sample: TreeNode;

// The recursive member must be an array of the same node type, not `+"`any`"+` and not `+"`unknown`"+`.
const leafName: string = sample.name;
const nestedName: string = sample.children[0].children[0].name;

// And the constraint must still bite.
// @ts-expect-error the recursive member is not a string
const wrong: string = sample.children;
`)
	assert.Equal(t, len(messages), 0, "expected no errors, got: "+strings.Join(messages, "; "))
}

// Skipping an accessor during a check that reports nothing is deferral only where the accessor
// genuinely cannot answer. Where it can, skipping it stops the constraint being checked at all, and
// the compiler accepts code it should reject. Real TypeScript 7.0.2 reports this; so must we.
func TestUnresolvedAccessorSkipDoesNotBlindConstraintChecking(t *testing.T) {
	t.Parallel()

	messages := checkSourceText(t, `
interface Schema<O> {
	readonly out: O;
}
type Shape = Record<string, Schema<any>>;
declare function object<S extends Shape>(shape: S): Schema<{ [K in keyof S]: S[K]["out"] }> & { shape: S };
declare const leaf: Schema<string>;

const flat = object({
	// An array of schemas is not a Schema, so the constraint on S is violated.
	get bad() {
		return [leaf];
	},
});
export const out = flat.out;
`)
	if len(messages) == 0 {
		t.Fatal("expected the constraint violation to be reported, got no errors")
	}
}

// Skipping a member that genuinely cannot be worked out yet postpones the constraint check; it does
// not waive it. Here the getter names the declaration being resolved, so the skip is correct, but the
// type it returns still violates the constraint and that has to be reported once the declaration has a
// type. Real TypeScript reports this too, via the circularity errors this change removes.
func TestPostponedConstraintCheckIsStillMade(t *testing.T) {
	t.Parallel()

	messages := checkSourceText(t, `
interface Schema<O> {
	readonly out: O;
}
type Shape = Record<string, Schema<any>>;
declare function object<S extends Shape>(shape: S): Schema<{ [K in keyof S]: S[K]["out"] }> & { shape: S };

const tree = object({
	// An array of schemas is not a Schema, even though it names the declaration being resolved.
	get bad() {
		return [tree];
	},
});
export const out = tree.out;
`)
	if len(messages) == 0 {
		t.Fatal("expected the postponed constraint violation to be reported, got no errors")
	}
}
