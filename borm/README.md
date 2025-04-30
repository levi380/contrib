
# borm

🏎️ 更好的ORM库 (Better ORM library that is simple, fast and self-mockable for Go)

# 目标
- 易用：SQL-Like（一把梭：One-Line-CRUD）
- KISS：保持小而美（不做大而全）
- 通用：支持struct，pb，map和基本类型
- 可测：支持自mock（因为参数作返回值，大部分mock框架不支持）
    - 非测试向的library不是好library
- As-Is：尽可能不作隐藏设定，防止误用
- 解决核心痛点：
   - 手撸SQL难免有错，组装数据太花时间
   - time.Time无法直接读写的问题
   - SQL函数结果无法直接Scan
   - db操作无法方便的Mock
   - QueryRow的sql.ErrNoRows问题
   - **直接替换系统自带Scanner，完整接管数据读取的类型转换**
- 核心原则：
   - 别像使用其他orm那样把一个表映射到一个model
   - （在borm里可以用Fields过滤器做到）
   - 尽量保持简单把一个操作映射一个model吧！
- 其他优点：
  - 更自然的where条件（仅在需要加括号时添加，对比gorm）
  - In操作接受各种类型slice，并且单元素时转成Equal操作
  - 从其他orm库迁移无需修改历史代码，无侵入性修改

# 特性矩阵

#### 下面是和一些主流orm库的对比

<table style="text-align: center">
   <tr>
      <td colspan="2">库</td>
      <td><a href="#">borm <strong>(me)</strong></a></td>
      <td><a href="https://github.com/jinzhu/gorm">gorm</a></td>
      <td><a href="https://github.com/go-xorm/xorm">xorm</a></td>
      <td>备注</td>
   </tr>
   <tr>
      <td rowspan="7">易用性</td>
      <td>无需指定类型</td>
      <td>:white_check_mark:</td>
      <td>:x:</td>
      <td>:x:</td>
      <td>borm在tag中无需低频的DDL</td>
   </tr>
   <tr>
      <td>无需指定model</td>
      <td>:white_check_mark:</td>
      <td>:x:</td>
      <td>:x:</td>
      <td>gorm/xorm改操作需提供“模版”</td>
   </tr>
   <tr>
      <td>无需指定主键</td>
      <td>:white_check_mark:</td>
      <td>:x:</td>
      <td>:x:</td>
      <td>gorm/xorm易误操作，如删/改全表</td>
   </tr>
   <tr>
      <td>学习成本低</td>
      <td>:white_check_mark:</td>
      <td>:x:</td>
      <td>:x:</td>
      <td>会SQL就会用borm</td>
   </tr>
   <tr>
      <td>可复用原生连接</td>
      <td>:white_check_mark:</td>
      <td>:x:</td>
      <td>:x:</td>
      <td>borm重构成本极小</td>
   </tr>
   <tr>
      <td>全类型转换</td>
      <td>:white_check_mark:</td>
      <td>maybe</td>
      <td>:x:</td>
      <td>杜绝类型转换的抛错</td>
   </tr>
   <tr>
      <td>复用查询命令</td>
      <td>:white_check_mark:</td>
      <td>:x:</td>
      <td>:x:</td>
      <td>borm批量和单条使用同一个函数</td>
   </tr>
   <tr>
      <td>可测试性</td>
      <td>自mock</td>
      <td>:white_check_mark:</td>
      <td>:x:</td>
      <td>:x:</td>
      <td>borm非常便于单元测试</td>
   </tr>
   <tr>
      <td rowspan="2">性能</td>
      <td>较原生耗时</td>
      <td><=1x</td>
      <td>2~3x</td>
      <td>2~3x</td>
      <td>xorm使用prepare模式会再慢2～3x</td>
   </tr>
   <tr>
      <td>反射</td>
      <td><a href="https://github.com/modern-go/reflect2">reflect2</a></td>
      <td>reflect</td>
      <td>reflect</td>
      <td>borm零使用ValueOf</td>
   </tr>
</table>

# 快速入门
   
2. 引入包
   ``` golang
   import b "contrib/borm"
   事物使用
   b.Transaction(db, func(tx *sql.Tx) error {

		d := TblBanner{
			Id:          "456",
			Title:       "我的标图",
			RedirectUrl: "http://www.163.com",
			Images:      "",
			Seq:         1,
			UrlType:     1,
			UpdatedName: "",
			UpdatedUid:  "11",
			UpdatedAt:   0,
			State:       "1",
			CreatedAt:   0,
		}

		t1 := b.Table(tx, "tbl_banner")
		_, err = t1.Insert(&d)

		fmt.Println("t1 Transaction err = ", err)

		a := TblAdmins{}
		t2 := b.Table(tx, "tbl_admins")
		_, err = t2.Insert(&a)
		fmt.Println("t2 Transaction err = ", err)
		return err
	})
    
   ```

3. 定义Table对象
   ``` golang
   t := b.Table(d.DB, "t_usr")

   t1 := b.Table(d.DB, "t_usr", ctx)
   ```

- `d.DB`是支持Exec/Query/QueryRow的数据库连接对象
- `t_usr`可以是表名，或者是嵌套查询语句
- `ctx`是需要传递的Context对象，默认不传为context.Background()

4. （可选）定义model对象
   ``` golang
   // Info 默认未设置borm tag的字段不会取
   type Info struct {
      ID   int64  `borm:"id"`
      Name string `borm:"name"`
      Tag  string `borm:"tag"`
   }

   // 调用t.UseNameWhenTagEmpty()，可以用未设置borm tag的字段名本身作为待获取的db字段
   ```

5. 执行操作

- **CRUD接口返回值为 (影响的条数，错误)**

- **类型`V`为`map[string]interface{}`的缩写形式，参考`gin.H`**

- 插入
   ``` golang
   // o可以是对象/slice/ptr slice
   n, err = t.Insert(&o)
   n, err = t.InsertIgnore(&o)
   n, err = t.ReplaceInto(&o)

   // 只插入部分字段（其他使用缺省）
   n, err = t.Insert(&o, b.Fields("name", "tag"))

   // 解决主键冲突
   n, err = t.Insert(&o, b.Fields("name", "tag"),
      b.OnDuplicateKeyUpdate(b.V{
         "name": "new_name",
         "age":  b.U("age+1"), // 使用b.U来处理非变量更新
      }))
   ```

- 查询
   ``` golang
   // o可以是对象/slice/ptr slice
   n, err := t.Select(&o, 
      b.Where("name = ?", name), 
      b.GroupBy("id"), 
      b.Having(b.Gt("id", 0)), 
      b.OrderBy("id", "name"), 
      b.Limit(1))

   // 使用基本类型+Fields获取条目数（n的值为1，因为结果只有1条）
   var cnt int64
   n, err = t.Select(&cnt, b.Fields("count(1)"), b.Where("name = ?", name))

   // 还可以支持数组
   var ids []int64
   n, err = t.Select(&ids, b.Fields("id"), b.Where("name = ?", name))

   // 可以强制索引
   n, err = t.Select(&ids, b.Fields("id"), b.ForceIndex("idx_xxx"), b.Where("name = ?", name))
   ```

- 更新
   ``` golang
   // o可以是对象/slice/ptr slice
   n, err = t.Update(&o, b.Where(b.Eq("id", id)))

   // 使用map更新
   n, err = t.Update(b.V{
         "name": "new_name",
         "tag":  "tag1,tag2,tag3",
         "age":  b.U("age+1"), // 使用b.U来处理非变量更新
      }, b.Where(b.Eq("id", id)), b.Limit(1))

   // 使用map更新部分字段
   n, err = t.Update(b.V{
         "name": "new_name",
         "tag":  "tag1,tag2,tag3",
      }, b.Fields("name"), b.Where(b.Eq("id", id)), b.Limit(1))

   n, err = t.Update(&o, b.Fields("name"), b.Where(b.Eq("id", id)), b.Limit(1))
   ```

- 删除
   ``` golang
   // 根据条件删除
   n, err = t.Delete(b.Where("name = ?", name))

   // 根据条件删除部分条数
   n, err = t.Delete(b.Where(b.Eq("id", id)), b.Limit(1))
   ```

- **可变条件**
   ``` golang
   conds := []interface{}{b.Cond("1=1")} // 防止空where条件
   if name != "" {
      conds = append(conds, b.Eq("name", name))
   }
   if id > 0 {
      conds = append(conds, b.Eq("id", id))
   }
   // 执行查询操作
   n, err := t.Select(&o, b.Where(conds...))
   ```

- **联表查询**
   ``` golang
   type Info struct {
      ID   int64  `borm:"t_usr.id"` // 字段定义加表名
      Name string `borm:"t_usr.name"`
      Tag  string `borm:"t_tag.tag"`
   }
   
   // 方法一
   t := b.Table(d.DB, "t_usr join t_tag on t_usr.id=t_tag.id") // 表名用join语句
   var o Info
   n, err := t.Select(&o, b.Where(b.Eq("t_usr.id", id))) // 条件加上表名

   // 方法二
   t = b.Table(d.DB, "t_usr") // 正常表名
   n, err = t.Select(&o, b.Join("join t_tag on t_usr.id=t_tag.id"), b.Where(b.Eq("t_usr.id", id))) // 条件需要加上表名
   ```

-  获取插入的自增id
   ``` golang
   // 首先需要数据库有一个自增ID的字段
   type Info struct {
      BormLastId int64 // 添加一个名为BormLastId的整型字段
      Name       string `borm:"name"`
      Age        string `borm:"age"`
   }

   o := Info{
      Name: "OrcaZ",
      Age:  30,
   }
   n, err = t.Insert(&o)

   id := o.BormLastId // 获取到插入的id
   ```
   
- 正在使用其他orm框架（新的接口先切过来吧）
   ``` golang
   // [gorm] db是一个*gorm.DB
   t := b.Table(db.DB(), "tbl")

   // [xorm] db是一个*xorm.EngineGroup
   t := b.Table(db.Master().DB().DB, "tbl")
   // or
   t := b.Table(db.Slave().DB().DB, "tbl")
   ```

# 其他细节

### Table的选项

|选项|说明|
|-|-|
|Debug|打印sql语句|
|Reuse|根据调用位置复用sql和存储方式|
|UseNameWhenTagEmpty|用未设置borm tag的字段名本身作为待获取的db字段|
|ToTimestamp|调用Insert时，使用时间戳，而非格式化字符串|

选项使用示例：
   ``` golang
   n, err = t.Debug().Insert(&o)

   n, err = t.ToTimestamp().Insert(&o)
   ```

### Where

|示例|说明|
|-|-|
|Where("id=? and name=?", id, name)|常规格式化版本|
|Where(Eq("id", id), Eq("name", name)...)|默认为and连接|
|Where(And(Eq("x", x), Eq("y", y), Or(Eq("x", x), Eq("y", y)...)...)|And & Or|

### 预置Where条件

|名称|示例|说明|
|-|-|-|
|逻辑与|And(...)|任意个参数，只接受下方的关系运算子|
|逻辑或|Or(...)|任意个参数，只接受下方的关系运算子|
|普通条件|Cond("id=?", id)|参数1为格式化字符串，后面跟占位参数|
|相等|Eq("id", id)|两个参数，id=?|
|不相等|Neq("id", id)|两个参数，id<>?|
|大于|Gt("id", id)|两个参数，id>?|
|大于等于|Gte("id", id)|两个参数，id>=?|
|小于|Lt("id", id)|两个参数，id<?|
|小于等于|Lte("id", id)|两个参数，id<=?|
|在...之间|Between("id", start, end)|三个参数，在start和end之间|
|近似|Like("name", "x%")|两个参数，name like "x%"|
|多值选择|In("id", ids)|两个参数，ids是基础类型的slice，slice只有1个元素会转化成Eq|

### GroupBy

|示例|说明|
|-|-|
|GroupBy("id", "name"...)|-|

### Having

|示例|说明|
|-|-|
|Having("id=? and name=?", id, name)|常规格式化版本|
|Having(Eq("id", id), Eq("name", name)...)|默认为and连接|
|Having(And(Eq("x", x), Eq("y", y), Or(Eq("x", x), Eq("y", y)...)...)|And & Or|

### OrderBy

|示例|说明|
|-|-|
|OrderBy("id desc", "name asc"...)|-|

### Limit

|示例|说明|
|-|-|
|Limit(1)|分页大小为1|
|Limit(0, 100)| limit ? offset ?|

### OnDuplicateKeyUpdate

|示例|说明|
|-|-|
|OnDuplicateKeyUpdate(V{"name": "new"})|解决主键冲突的更新|

### ForceIndex

|示例|说明|
|-|-|
|ForceIndex("idx_biz_id")|解决索引选择性差的问题|

