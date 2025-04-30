package borm

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path"
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	"github.com/modern-go/reflect2"
)

// Select .

func fieldEscapeSQL(sb *strings.Builder, field string) {
	if field == "" {
		return
	}
	sb.WriteString(field)
}

func (t *BormTable) DeleteSQL() (int, error) {

	if len(t.Args) <= 0 {
		return 0, errors.New("argument 1 cannot be omitted")
	}

	var (
		sb       strings.Builder
		stmtArgs []interface{}
	)
	sb.WriteString("delete from ")
	fieldEscapeSQL(&sb, t.Name)

	for _, arg := range t.Args {
		arg.BuildString(&sb)
		//arg.BuildArgs(&stmtArgs)
	}

	if t.Cfg.Debug {
		log.Println(sb.String())
		//fmt.Println(sb.String(), stmtArgs)
	}

	t.Args = []BormItem{}

	res, err := t.DB.ExecContext(t.ctx, sb.String(), stmtArgs...)
	if err != nil {
		return 0, err
	}

	row, _ := res.RowsAffected()
	return int(row), nil
}

func (t *BormTable) SelectSQL(res interface{}) (int, error) {

	if len(t.Args) <= 0 {
		return 0, errors.New("argument 2 cannot be omitted")
	}

	var (
		rt         = reflect2.TypeOf(res)
		isArray    bool
		isPtrArray bool
		rtElem     = rt

		item     *DataBindingItem
		stmtArgs []interface{}
	)

	switch rt.Kind() {
	case reflect.Ptr:
		rt = rt.(reflect2.PtrType).Elem()
		rtElem = rt
		if rt.Kind() == reflect.Slice {
			rtElem = rt.(reflect2.ListType).Elem()
			isArray = true

			if rtElem.Kind() == reflect.Ptr {
				rtElem = rtElem.(reflect2.PtrType).Elem()
				isPtrArray = true
			}
		}
		// case reflect.Map:
		// TODO
	default:
		return 0, errors.New("argument 2 should be map or ptr")
	}

	if t.Cfg.Reuse {
		_, fileName, line, _ := runtime.Caller(1)
		item = loadFromCache(fileName, line)
	}

	if item != nil {
		// struct类型
		if rtElem.Kind() == reflect.Struct {
			if t.Args[0].Type() == _fields {
				t.Args = t.Args[1:]
			}
			// map类型
			// } else if rt.Kind() == reflect.Map {
			// TODO
			// 其他类型
		} else {
			t.Args = t.Args[1:]
		}

		for _, arg := range t.Args {
			arg.BuildArgs(&stmtArgs)
		}
	} else {
		item = &DataBindingItem{Type: rtElem}

		var sb strings.Builder
		sb.WriteString("select ")

		if isArray {
			item.Elem = rtElem.New()
		} else {
			item.Elem = res
		}

		// struct类型
		if rtElem.Kind() == reflect.Struct {
			s := rtElem.(reflect2.StructType)

			if t.Args[0].Type() == _fields {
				m := t.getStructFieldMap(s)

				for _, field := range t.Args[0].(*fieldsItem).Fields {
					chunk := strings.SplitN(field, " as ", 3)
					if len(chunk) == 2 {
						field = strings.TrimSpace(chunk[1])
					}
					f, ok := m[field]
					if ok {
						item.Cols = append(item.Cols, &scanner{
							Type: f.Type(),
							Val:  f.UnsafeGet(reflect2.PtrOf(item.Elem)),
						})
					} else {
						fmt.Println("BormTable Select m = ", m)
						fmt.Println("BormTable Select field not found = ", field)
					}
				}

				(t.Args[0]).BuildSQL(&sb)
				t.Args = t.Args[1:]

			} else {
				for i := 0; i < s.NumField(); i++ {
					f := s.Field(i)
					ft := f.Tag().Get("borm")

					if !t.Cfg.UseNameWhenTagEmpty && ft == "" {
						continue
					}

					if len(item.Cols) > 0 {
						sb.WriteString(",")
					}

					if ft == "" {
						fieldEscapeSQL(&sb, f.Name())
					} else {
						fieldEscapeSQL(&sb, ft)
					}

					item.Cols = append(item.Cols, &scanner{
						Type: f.Type(),
						Val:  f.UnsafeGet(reflect2.PtrOf(item.Elem)),
					})
				}
			}
			// map类型
			// } else if rt.Kind() == reflect.Map {
			// TODO
			// 其他类型
		} else {
			// 必须有fields且为1
			if t.Args[0].Type() != _fields {
				return 0, errors.New("argument 3 need ONE Fields(\"name\") with ONE field")
			}

			fi := t.Args[0].(*fieldsItem)
			if len(fi.Fields) < 1 {
				return 0, errors.New("too few fields")
			}

			item.Cols = append(item.Cols, &scanner{
				Type: rtElem,
				Val:  reflect2.PtrOf(item.Elem),
			})

			fieldEscapeSQL(&sb, fi.Fields[0])
			t.Args = t.Args[1:]
		}

		sb.WriteString(" from ")

		fieldEscapeSQL(&sb, t.Name)

		for _, arg := range t.Args {
			arg.BuildString(&sb)
			//arg.BuildArgs(&stmtArgs)
		}

		item.SQL = sb.String()

		if t.Cfg.Reuse {
			_, fileName, line, _ := runtime.Caller(1)
			storeToCache(fileName, line, item)
		}
	}

	//fmt.Println("isPtrArray = ", isPtrArray)
	if t.Cfg.Debug {
		fmt.Println("item.SQL = ", item.SQL)
		//query := fmt.Sprintf(strings.ReplaceAll(item.SQL, "?", "%v"), stmtArgs...)
		//log.Println(query)
		//fmt.Println(item.SQL, stmtArgs)
	}
	t.Args = []BormItem{}
	if !isArray {
		// fire
		err := t.DB.QueryRowContext(t.ctx, item.SQL, stmtArgs...).Scan(item.Cols...)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, nil
			}
			return 0, err
		}
		return 1, err
	}

	// fire
	rows, err := t.DB.QueryContext(t.ctx, item.SQL, stmtArgs...)
	if err != nil {
		return 0, err
	}

	count := 0
	for rows.Next() {
		err = rows.Scan(item.Cols...)
		if err != nil {
			break
		}

		if isPtrArray {
			copyElem := rtElem.UnsafeNew()
			rtElem.UnsafeSet(copyElem, reflect2.PtrOf(item.Elem))
			rt.(reflect2.SliceType).UnsafeAppend(reflect2.PtrOf(res), unsafe.Pointer(&copyElem))
		} else {
			rt.(reflect2.SliceType).UnsafeAppend(reflect2.PtrOf(res), reflect2.PtrOf(item.Elem))
		}
		count++
	}
	rows.Close()
	return count, err
}

func (t *BormTable) inputValues(sbTmp *strings.Builder, cols []reflect2.StructField, rtPtr, s reflect2.Type, ptr bool, x unsafe.Pointer) {

	ll := len(cols) - 1
	for j, col := range cols {
		var val interface{}
		if ptr {
			val = col.Get(rtPtr.UnsafeIndirect(x))
		} else {
			val = col.Get(s.PackEFace(x))
		}
		switch v := val.(type) {
		case *int:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *int8:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *int16:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *int32:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *int64:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *uint:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *uint8:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *uint16:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *uint32:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *uint64:
			sbTmp.WriteString(fmt.Sprintf("%d", *v))
		case *float64:
			sbTmp.WriteString(fmt.Sprintf("%.2f", *v))
		case *string:
			sbTmp.WriteString("'")
			sbTmp.WriteString(*v)
			sbTmp.WriteString("'")
		default:
			sbTmp.WriteString(fmt.Sprintf("%v", v))
		}
		if j < ll {
			sbTmp.WriteString(",")
		}

		/*
			// 时间类型特殊处理
			if col.Type().String() == "time.Time" {
				if t.Cfg.ToTimestamp {
					v = v.(*time.Time).Unix()
				} else {
					v = v.(*time.Time).Format(_timeLayout)
				}
			}
		*/
		//*stmtArgs = append(*stmtArgs, v)
	}
	sbTmp.WriteString(")")
}

func (t *BormTable) UpdateSQL(obj interface{}) (int, error) {

	if len(t.Args) <= 0 {
		return 0, errors.New("argument 2 cannot be omitted")
	}

	var sb strings.Builder
	sb.WriteString("update ")
	fieldEscapeSQL(&sb, t.Name)
	sb.WriteString(" set ")

	var stmtArgs []interface{}

	if m, ok := obj.(V); ok {
		if t.Args[0].Type() == _fields {
			argCnt := 0
			for _, field := range t.Args[0].(*fieldsItem).Fields {
				v := m[field]
				if v != nil {
					if argCnt > 0 {
						sb.WriteString(",")
					}
					fieldEscapeSQL(&sb, field)
					if s, ok := v.(U); ok {
						sb.WriteString("=")
						sb.WriteString(string(s))
					} else {
						sb.WriteString("=?")
						stmtArgs = append(stmtArgs, v)
					}
					fmt.Println("field = ", field, " v = ", v)
					argCnt++
				}
			}

			t.Args = t.Args[1:]

		} else {
			argCnt := 0
			for k, v := range m {
				if argCnt > 0 {
					sb.WriteString(",")
				}
				fieldEscapeSQL(&sb, k)
				if s, ok := v.(U); ok {
					sb.WriteString("=")
					sb.WriteString(string(s))
				} else {
					sb.WriteString("=")

					switch val := v.(type) {
					case int, int32, int64, uint, uint32, uint64:
						sb.WriteString(fmt.Sprintf("%d", v))
					case float32, float64:
						sb.WriteString(fmt.Sprintf("%f", v))
					case string:
						sb.WriteString("'")
						sb.WriteString(val)
						sb.WriteString("'")
					default:
						fmt.Printf("Unknown type: %T, value: %v\n", v, v)
					}
				}
				argCnt++
			}
		}
	} else {
		rt := reflect2.TypeOf(obj)

		switch rt.Kind() {
		case reflect.Ptr:
			rt = rt.(reflect2.PtrType).Elem()
		// case reflect.Map:
		// TODO
		default:
			return 0, errors.New("argument 2 should be map or ptr")
		}

		var cols []reflect2.StructField

		// Fields or None
		// struct类型
		if rt.Kind() != reflect.Struct {
			return 0, errors.New("non-structure type not supported yet")
		}

		// Fields or KeyVals or None
		s := rt.(reflect2.StructType)
		if t.Args[0].Type() == _fields {
			m := t.getStructFieldMap(s)

			fmt.Println("111")
			for i, field := range t.Args[0].(*fieldsItem).Fields {
				f := m[field]
				if f != nil {
					cols = append(cols, f)
				}

				if i > 0 {
					sb.WriteString(",")
				}
				fieldEscapeSQL(&sb, field)
				sb.WriteString("=?")
			}

			t.Args = t.Args[1:]

		} else {
			fmt.Println("222")
			for i := 0; i < s.NumField(); i++ {
				f := s.Field(i)
				ft := f.Tag().Get("borm")

				if !t.Cfg.UseNameWhenTagEmpty && ft == "" {
					continue
				}

				if len(cols) > 0 {
					sb.WriteString(",")
				}

				if ft == "" {
					fieldEscapeSQL(&sb, f.Name())
					sb.WriteString("=")
				} else {

					fieldEscapeSQL(&sb, ft)
					sb.WriteString("=")
				}

				vv := f.Get(s.PackEFace(reflect2.PtrOf(obj)))

				switch val := vv.(type) {
				case *int:
					sb.WriteString(fmt.Sprintf("%d", *vv.(*int)))
				case *int32:
					sb.WriteString(fmt.Sprintf("%d", *vv.(*int32)))
				case *int64:
					sb.WriteString(fmt.Sprintf("%d", *vv.(*int64)))
				case *uint:
					sb.WriteString(fmt.Sprintf("%d", *vv.(*uint)))
				case *uint32:
					sb.WriteString(fmt.Sprintf("%d", *vv.(*uint32)))
				case *uint64:
					sb.WriteString(fmt.Sprintf("%d", *vv.(*uint64)))
				case *float32:
					sb.WriteString(fmt.Sprintf("%f", *vv.(*float32)))
				case *float64:
					sb.WriteString(fmt.Sprintf("%f", *vv.(*float64)))
				case *string:
					sb.WriteString("'")
					sb.WriteString(*val)
					sb.WriteString("'")
				}

				sb.WriteString(",")
			}
		}

		//t.inputArgs(&stmtArgs, cols, rt, s, false, reflect2.PtrOf(obj))
	}

	for _, arg := range t.Args {
		arg.BuildString(&sb)
		//arg.BuildArgs(&stmtArgs)
	}

	if t.Cfg.Debug {
		log.Println(sb.String())
		//fmt.Println(sb.String(), stmtArgs)
	}
	t.Args = []BormItem{}
	res, err := t.DB.ExecContext(t.ctx, sb.String(), stmtArgs...)
	if err != nil {
		return 0, err
	}

	row, _ := res.RowsAffected()
	return int(row), nil
}

func (t *BormTable) ReplaceIntoSQL(objs interface{}, args ...BormItem) (int, error) {
	if config.Mock {
		pc, fileName, _, _ := runtime.Caller(1)
		if ok, _, n, e := checkMock(t.Name, "ReplaceInto", runtime.FuncForPC(pc).Name(), fileName, path.Dir(fileName)); ok {
			return n, e
		}
	}

	return t.insertPack("replace into ", objs, args)
}

func (t *BormTable) InsertSQL(objs interface{}, args ...BormItem) (int, error) {
	if config.Mock {
		pc, fileName, _, _ := runtime.Caller(1)
		if ok, _, n, e := checkMock(t.Name, "Insert", runtime.FuncForPC(pc).Name(), fileName, path.Dir(fileName)); ok {
			return n, e
		}
	}

	return t.insertPack("insert into ", objs, args)
}

func (t *BormTable) insertPack(prefix string, objs interface{}, args []BormItem) (int, error) {
	var (
		rt         = reflect2.TypeOf(objs)
		isArray    bool
		isPtrArray bool
		rtPtr      reflect2.Type
		rtElem     = rt

		sb   strings.Builder
		cols []reflect2.StructField
	)

	//sb.WriteString("insert into ")
	sb.WriteString(prefix)

	fieldEscapeSQL(&sb, t.Name)

	sb.WriteString(" (")

	switch rt.Kind() {
	case reflect.Ptr:
		rt = rt.(reflect2.PtrType).Elem()
		rtElem = rt
		if rt.Kind() == reflect.Slice {
			rtElem = rtElem.(reflect2.ListType).Elem()
			isArray = true

			if rtElem.Kind() == reflect.Ptr {
				rtPtr = rtElem
				rtElem = rtElem.(reflect2.PtrType).Elem()
				isPtrArray = true
			}
		}
	// case reflect.Map:
	// TODO
	default:
		return 0, errors.New("argument 2 should be map or ptr")
	}

	// Fields or None
	// struct类型
	if rtElem.Kind() != reflect.Struct {
		return 0, errors.New("non-structure type not supported yet")
	}

	s := rtElem.(reflect2.StructType)

	if len(args) > 0 && args[0].Type() == _fields {
		m := t.getStructFieldMap(s)

		for _, field := range args[0].(*fieldsItem).Fields {
			f := m[field]
			if f != nil {
				cols = append(cols, f)
			}
		}

		(args[0]).BuildSQL(&sb)
		args = args[1:]

	} else {
		for i := 0; i < s.NumField(); i++ {
			f := s.Field(i)
			ft := f.Tag().Get("borm")

			if !t.Cfg.UseNameWhenTagEmpty && ft == "" {
				continue
			}

			if len(cols) > 0 {
				sb.WriteString(",")
			}

			if ft == "" {
				fieldEscape(&sb, f.Name())
			} else {
				fieldEscape(&sb, ft)
			}

			cols = append(cols, f)
		}
	}

	sb.WriteString(") values ")
	sb.WriteString("(")
	if isArray {
		// 数组
		for i := 0; i < rt.(reflect2.SliceType).UnsafeLengthOf(reflect2.PtrOf(objs)); i++ {
			if i > 0 {
				sb.WriteString(",(")
			}
			//sb.WriteString(sbTmp.String())
			t.inputValues(&sb, cols, rtPtr, s, isPtrArray, rt.(reflect2.ListType).UnsafeGetIndex(reflect2.PtrOf(objs), i))
		}
	} else {
		// 普通元素
		t.inputValues(&sb, cols, rtPtr, s, false, reflect2.PtrOf(objs))
	}

	if t.Cfg.Debug {
		log.Println(sb.String())
	}

	//if 1 == 1 {
	//	return 0, nil
	//}
	res, err := t.DB.ExecContext(t.ctx, sb.String())
	if err != nil {
		return 0, err
	}

	if !isArray {
		if f := s.FieldByName("BormLastId"); f != nil {
			id, _ := res.LastInsertId()
			f.UnsafeSet(reflect2.PtrOf(objs), reflect2.PtrOf(id))
		}
	}

	row, _ := res.RowsAffected()
	return int(row), nil
}

// InsertAndGetID 插入单条数据并返回插入的ID
func (t *BormTable) InsertAndGetID(obj interface{}, args ...BormItem) (int64, error) {
	if config.Mock {
		pc, fileName, _, _ := runtime.Caller(1)
		if ok, _, n, e := checkMock(t.Name, "Insert", runtime.FuncForPC(pc).Name(), fileName, path.Dir(fileName)); ok {
			return int64(n), e
		}
	}

	return t.insertPackToGetId("insert into ", obj, args)
}

func (t *BormTable) insertPackToGetId(prefix string, obj interface{}, args []BormItem) (int64, error) {
	var (
		rt = reflect2.TypeOf(obj)
		// isArray    bool
		// isPtrArray bool
		rtPtr  reflect2.Type
		rtElem = rt

		sb   strings.Builder
		cols []reflect2.StructField
	)

	sb.WriteString(prefix)
	fieldEscapeSQL(&sb, t.Name)
	sb.WriteString(" (")

	switch rt.Kind() {
	case reflect.Ptr:
		rt = rt.(reflect2.PtrType).Elem()
		rtElem = rt
		if rt.Kind() == reflect.Slice {
			return 0, errors.New("array insert not supported in InsertAndGetID")
		}
	default:
		return 0, errors.New("argument 2 should be map or ptr")
	}

	if rtElem.Kind() != reflect.Struct {
		return 0, errors.New("non-structure type not supported yet")
	}

	s := rtElem.(reflect2.StructType)

	if len(args) > 0 && args[0].Type() == _fields {
		m := t.getStructFieldMap(s)
		for _, field := range args[0].(*fieldsItem).Fields {
			f := m[field]
			if f != nil {
				cols = append(cols, f)
			}
		}
		(args[0]).BuildSQL(&sb)
		args = args[1:]
	} else {
		for i := 0; i < s.NumField(); i++ {
			f := s.Field(i)
			ft := f.Tag().Get("borm")

			if !t.Cfg.UseNameWhenTagEmpty && ft == "" {
				continue
			}

			if len(cols) > 0 {
				sb.WriteString(",")
			}

			if ft == "" {
				fieldEscape(&sb, f.Name())
			} else {
				fieldEscape(&sb, ft)
			}

			cols = append(cols, f)
		}
	}

	sb.WriteString(") values (")
	t.inputValues(&sb, cols, rtPtr, s, false, reflect2.PtrOf(obj))

	if t.Cfg.Debug {
		log.Println(sb.String())
	}

	res, err := t.DB.ExecContext(t.ctx, sb.String())
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	if f := s.FieldByName("BormLastId"); f != nil {
		f.UnsafeSet(reflect2.PtrOf(obj), reflect2.PtrOf(lastID))
	}

	return lastID, nil
}

// BuildBatchInsertSQL 构建批量插入SQL语句和参数
func BuildBatchInsertSQL(tableName string, data []map[string]interface{}) (string, []interface{}) {
	if len(data) == 0 {
		return "", nil
	}

	// 获取所有字段名
	var fields []string
	fieldMap := make(map[string]bool)

	// 首先收集所有可能的字段
	for _, item := range data {
		for field := range item {
			if !fieldMap[field] {
				fieldMap[field] = true
				fields = append(fields, field)
			}
		}
	}

	// 构建SQL语句
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(tableName)
	sb.WriteString(" (")

	// 添加字段名
	for i, field := range fields {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(field)
	}

	sb.WriteString(") VALUES ")

	// 添加参数占位符
	var args []interface{}
	for i, item := range data {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString("(")
		for j, field := range fields {
			if j > 0 {
				sb.WriteString(", ")
			}

			// 如果该条数据有这个字段，使用其值；否则使用NULL
			if val, ok := item[field]; ok {
				sb.WriteString("?")
				args = append(args, val)
			} else {
				sb.WriteString("NULL")
			}
		}
		sb.WriteString(")")
	}

	return sb.String(), args
}
