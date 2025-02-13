package filter

import "gorm.io/gen"

type Filter interface {
	// SELECT * FROM @@table WHERE id = @id LIMIT 1
	GetByID(id uint64) gen.T
	// SELECT * FROM @@table
	// {{where}}
	//   {{if strings.HasPrefix(val, "%") && strings.HasSuffix(val, "%")}}
	//     @@col LIKE concat('%', TRIM(BOTH '%' FROM @val), '%')
	//   {{else if strings.HasPrefix(val, "%")}}
	//     @@col LIKE concat('%', TRIM(BOTH '%' FROM @val))
	//   {{else if strings.HasSuffix(val, "%")}}
	//     @@col LIKE concat(TRIM(BOTH '%' FROM @val), '%')
	//   {{else}}
	//     @@col = @val
	//   {{end}}
	// {{end}}
	// LIMIT 1
	GetByCol(col, val string) gen.T
	// SELECT * FROM @@table
	// {{if len(cols) == len(vals)}}
	//  {{where}}
	//     {{for colIndex, col := range cols}}
	//       {{for valIndex, val := range vals}}
	//         {{if colIndex == valIndex}}
	//           {{if strings.HasPrefix(val, "%") && strings.HasSuffix(val, "%")}}
	//             @@col LIKE concat('%', TRIM(BOTH '%' FROM @val), '%') AND
	//           {{else if strings.HasPrefix(val, "%")}}
	//             @@col LIKE concat('%', TRIM(BOTH '%' FROM @val)) AND
	//           {{else if strings.HasSuffix(val, "%")}}
	//             @@col LIKE concat(TRIM(BOTH '%' FROM @val), '%') AND
	//           {{else}}
	//             @@col = @val AND
	//           {{end}}
	//         {{end}}
	//       {{end}}
	//     {{end}}
	//   {{end}}
	// {{end}}
	// LIMIT 1
	GetByCols(cols, vals []string) gen.T
	// SELECT * FROM @@table
	// {{where}}
	//   {{if strings.HasPrefix(val, "%") && strings.HasSuffix(val, "%")}}
	//     @@col LIKE concat('%', TRIM(BOTH '%' FROM @val), '%')
	//   {{else if strings.HasPrefix(val, "%")}}
	//     @@col LIKE concat('%', TRIM(BOTH '%' FROM @val))
	//   {{else if strings.HasSuffix(val, "%")}}
	//     @@col LIKE concat(TRIM(BOTH '%' FROM @val), '%')
	//   {{else}}
	//     @@col = @val
	//   {{end}}
	// {{end}}
	FindByCol(col, val string) []gen.T
	// SELECT * FROM @@table
	// {{if len(cols) == len(vals)}}
	//  {{where}}
	//     {{for colIndex, col := range cols}}
	//       {{for valIndex, val := range vals}}
	//         {{if colIndex == valIndex}}
	//           {{if strings.HasPrefix(val, "%") && strings.HasSuffix(val, "%")}}
	//             @@col LIKE concat('%', TRIM(BOTH '%' FROM @val), '%') AND
	//           {{else if strings.HasPrefix(val, "%")}}
	//             @@col LIKE concat('%', TRIM(BOTH '%' FROM @val)) AND
	//           {{else if strings.HasSuffix(val, "%")}}
	//             @@col LIKE concat(TRIM(BOTH '%' FROM @val), '%') AND
	//           {{else}}
	//             @@col = @val AND
	//           {{end}}
	//         {{end}}
	//       {{end}}
	//     {{end}}
	//   {{end}}
	// {{end}}
	FindByCols(cols, vals []string) []gen.T
}
