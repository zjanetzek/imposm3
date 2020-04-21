package postgis

import (
	"fmt"

	"github.com/omniscale/imposm3/log"
)

type ColumnType interface {
	Name() string
	PrepareInsertSQL(i int,
		spec *TableSpec) string
	GeneralizeSQL(colSpec *ColumnSpec, spec *GeneralizedTableSpec) string
}

type simpleColumnType struct {
	name string
}

func (t *simpleColumnType) Name() string {
	return t.name
}

func (t *simpleColumnType) PrepareInsertSQL(i int, spec *TableSpec) string {
	return fmt.Sprintf("$%d", i)
}

func (t *simpleColumnType) GeneralizeSQL(colSpec *ColumnSpec, spec *GeneralizedTableSpec) string {
	return "\"" + colSpec.Name + "\""
}

type geometryType struct {
	name string
}

func (t *geometryType) Name() string {
	return t.name
}

func (t *geometryType) PrepareInsertSQL(i int, spec *TableSpec) string {
	return fmt.Sprintf("$%d::Geometry",
		i,
	)
}

func (t *geometryType) GeneralizeSQL(colSpec *ColumnSpec, spec *GeneralizedTableSpec) string {
	if (spec.Simplify == ``) {
		log.Printf("[warn] validated_geometry %s", fmt.Sprintf(`ST_Simplify("%s", %f) as "%s"`,
			colSpec.Name, spec.Tolerance, colSpec.Name,
		))
		return fmt.Sprintf(`ST_Simplify("%s", %f) as "%s"`,
			colSpec.Name, spec.Tolerance, colSpec.Name,
		)
	}
	log.Printf("[warn] validated_geometry %s", fmt.Sprintf(spec.Simplify + `("%s", %f) as "%s"`,
		colSpec.Name, spec.Tolerance, colSpec.Name,
	))
	return fmt.Sprintf(spec.Simplify + `("%s", %f) as "%s"`,
		colSpec.Name, spec.Tolerance, colSpec.Name,
	)
}

type validatedGeometryType struct {
	geometryType
}

func (t *validatedGeometryType) GeneralizeSQL(colSpec *ColumnSpec, spec *GeneralizedTableSpec) string {
	if spec.Source.GeometryType != "polygon" {
		// TODO return warning earlier
		log.Printf("[warn] validated_geometry column returns polygon geometries for %s", spec.FullName)
	}
	if (spec.SimplifyValidated == ``) {
		log.Printf("[warn] validated_geometry %s", fmt.Sprintf(`ST_Buffer(ST_SimplifyPreserveTopology("%s", %f), 0) as "%s"`,
			colSpec.Name, spec.Tolerance, colSpec.Name,
		))
		return fmt.Sprintf(`ST_Buffer(ST_SimplifyPreserveTopology("%s", %f), 0) as "%s"`,
			colSpec.Name, spec.Tolerance, colSpec.Name,
		)
	}
	log.Printf("[warn] validated_geometry %s", fmt.Sprintf(spec.SimplifyValidated + `("%s", %f) as "%s"`,
		colSpec.Name, spec.Tolerance, colSpec.Name,
	))
	return fmt.Sprintf(spec.SimplifyValidated + `("%s", %f) as "%s"`,
		colSpec.Name, spec.Tolerance, colSpec.Name,
	)
}

var pgTypes map[string]ColumnType

func init() {
	pgTypes = map[string]ColumnType{
		"string":             &simpleColumnType{"VARCHAR"},
		"bool":               &simpleColumnType{"BOOL"},
		"int8":               &simpleColumnType{"SMALLINT"},
		"int32":              &simpleColumnType{"INT"},
		"int64":              &simpleColumnType{"BIGINT"},
		"float32":            &simpleColumnType{"REAL"},
		"hstore_string":      &simpleColumnType{"HSTORE"},
		"geometry":           &geometryType{"GEOMETRY"},
		"validated_geometry": &validatedGeometryType{geometryType{"GEOMETRY"}},
	}
}
