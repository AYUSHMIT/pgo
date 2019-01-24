package pgo.model.mpcal;

import pgo.TODO;
import pgo.formatters.IndentingWriter;
import pgo.formatters.PlusCalNodeFormattingVisitor;
import pgo.model.pcal.PlusCalStatement;

import java.io.IOException;
import java.util.stream.Collectors;

public class ModularPlusCalNodeFormattingVisitor extends ModularPlusCalNodeVisitor<Void, IOException> {
	private IndentingWriter out;

	public ModularPlusCalNodeFormattingVisitor(IndentingWriter out) {
		this.out = out;
	}

	@Override
	public Void visit(ModularPlusCalBlock modularPlusCalBlock) throws IOException {
		throw new TODO();
	}

	@Override
	public Void visit(ModularPlusCalArchetype modularPlusCalArchetype) throws IOException {
		out.write("archetype ");
		out.write(modularPlusCalArchetype.getName());
		out.write("(");
		out.write(modularPlusCalArchetype
				.getParams()
				.stream()
				.map(arg -> (arg.isRef() ? "ref " : "") + arg.getName().getValue())
				.collect(Collectors.joining(", ")));
		out.write(")");
		if (modularPlusCalArchetype.getVariables().isEmpty()) {
			out.write(" ");
		} else {
			out.write("variables ");
			out.write(modularPlusCalArchetype
					.getVariables()
					.stream()
					.map(v -> v.getName() + (v.isSet() ? " \\in " : " = ") + v.getValue().toString())
					.collect(Collectors.joining(", ")));
			out.write(";");
			out.newLine();
		}
		out.write("{");
		// TODO write body
		out.write("}");
		return null;
	}

	@Override
	public Void visit(ModularPlusCalInstance modularPlusCalInstance) throws IOException {
		throw new TODO();
	}

	@Override
	public Void visit(ModularPlusCalMappingMacro modularPlusCalMappingMacro) throws IOException {
		out.write("mapping macro ");
		out.write(modularPlusCalMappingMacro.getName());
		out.write("{");

		out.write("read {");
		for (PlusCalStatement s : modularPlusCalMappingMacro.getReadBody()) {
			s.accept(new PlusCalNodeFormattingVisitor(out));
		}
		out.write("}");

		out.write("write {");
		for (PlusCalStatement s : modularPlusCalMappingMacro.getWriteBody()) {
			s.accept(new PlusCalNodeFormattingVisitor(out));
		}
		out.write("}");

		out.write("}");
		return null;
	}
}
