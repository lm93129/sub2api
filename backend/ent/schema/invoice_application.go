package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InvoiceApplication holds the schema definition for the InvoiceApplication entity.
//
// 开票申请：用户勾选若干笔已充值的余额订单提交开票申请，管理员审核并登记发票信息。
//
// 删除策略：硬删除
// 申请单是财务凭证，其生命周期完全由 status 字段驱动（PENDING → APPROVED → ISSUED
// 或 PENDING → REJECTED / CANCELLED），不提供删除入口，因此不需要软删除。
//
// 关联策略：不建外键
// user_id 与 invoice_application_items.order_id 均为纯整数列，不建外键：
//   - payment_orders 使用硬删除，建外键会阻断订单删除，改变既有行为；
//   - 用户被删除后开票凭证仍需留存以备审计。
//
// 因此用户邮箱、用户名与订单号、实付金额均在本表/明细表冗余存储快照。
type InvoiceApplication struct {
	ent.Schema
}

func (InvoiceApplication) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "invoice_applications"},
	}
}

func (InvoiceApplication) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (InvoiceApplication) Fields() []ent.Field {
	return []ent.Field{
		// 申请用户（冗余邮箱与用户名，避免关联查询；用户删除后凭证仍可读）
		field.Int64("user_id").
			Comment("申请用户ID"),
		field.String("user_email").
			MaxLen(255).
			Default("").
			Comment("申请用户邮箱（快照）"),
		field.String("user_name").
			MaxLen(100).
			Default("").
			Comment("申请用户名称（快照）"),

		// 状态：PENDING, APPROVED, ISSUED, REJECTED, CANCELLED
		field.String("status").
			MaxLen(20).
			Default("PENDING").
			Comment("状态: PENDING待审核, APPROVED已审核待开票, ISSUED已开票, REJECTED已驳回, CANCELLED用户已取消"),

		// 发票类型：electron_normal 电子普通发票, electron_special 电子专用发票
		field.String("invoice_type").
			MaxLen(30).
			Default("electron_normal").
			Comment("发票类型: electron_normal电子普通发票, electron_special电子专用发票"),
		// 抬头类型：personal 个人, company 企业
		field.String("title_type").
			MaxLen(20).
			Default("personal").
			Comment("抬头类型: personal个人, company企业"),
		field.String("title").
			MaxLen(200).
			NotEmpty().
			Comment("发票抬头（单位名称或个人姓名）"),
		field.String("tax_number").
			MaxLen(64).
			Default("").
			Comment("纳税人识别号（统一社会信用代码），个人抬头可为空"),
		field.String("registered_address").
			MaxLen(255).
			Default("").
			Comment("注册地址"),
		field.String("registered_phone").
			MaxLen(50).
			Default("").
			Comment("注册电话"),
		field.String("bank_name").
			MaxLen(200).
			Default("").
			Comment("开户银行"),
		field.String("bank_account").
			MaxLen(100).
			Default("").
			Comment("银行账号"),
		field.String("email").
			MaxLen(255).
			NotEmpty().
			Comment("接收发票的邮箱"),
		field.String("remark").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default("").
			Comment("用户备注"),

		// 金额（人民币，decimal(20,2)）
		field.Float("total_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0).
			Comment("申请金额合计（所选充值记录实付金额之和）"),
		field.Float("invoice_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Optional().
			Nillable().
			Comment("实际开票金额（管理员登记，可与申请金额不同）"),

		// 开票信息（管理员登记）
		field.String("invoice_number").
			MaxLen(64).
			Default("").
			Comment("发票号码"),
		field.String("invoice_code").
			MaxLen(64).
			Default("").
			Comment("发票代码"),
		field.Time("issued_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("开票日期"),

		// 审核信息
		field.Int64("reviewed_by").
			Optional().
			Nillable().
			Comment("审核人用户ID（管理员）"),
		field.Time("reviewed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("审核时间"),
		field.String("rejected_reason").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default("").
			Comment("驳回原因"),
		field.String("admin_remark").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default("").
			Comment("管理员备注"),
	}
}

func (InvoiceApplication) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", InvoiceApplicationItem.Type),
	}
}

func (InvoiceApplication) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "status"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}

// InvoiceApplicationItem holds the schema definition for the InvoiceApplicationItem entity.
//
// 开票申请明细：记录一条申请占用了哪一笔充值订单。
//
// 占用语义：
//   - released_at IS NULL 表示该充值记录仍被本申请占用；
//   - 申请被驳回或用户取消时写入 released_at，释放占用，记录保留以供审计；
//   - order_id 上的部分唯一索引（WHERE released_at IS NULL）保证一笔充值记录
//     至多被一条未释放的申请占用，并发提交也无法突破该约束。
//
// 关联策略：不建外键
// payment_orders 使用硬删除，建外键会阻断订单删除，因此 order_id 为纯整数列，
// 订单号、实付金额与支付时间均冗余存储快照。
type InvoiceApplicationItem struct {
	ent.Schema
}

func (InvoiceApplicationItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "invoice_application_items"},
	}
}

func (InvoiceApplicationItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (InvoiceApplicationItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("application_id").
			Comment("所属开票申请ID"),
		field.Int64("order_id").
			Comment("被占用的充值订单ID（payment_orders.id）"),
		field.String("order_no").
			MaxLen(64).
			Default("").
			Comment("充值订单号（快照）"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0).
			Comment("该订单实付金额（快照，人民币）"),
		field.String("currency").
			MaxLen(10).
			Default("CNY").
			Comment("支付币种快照"),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("支付时间（快照）"),
		field.Time("released_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("占用释放时间，为空表示仍被占用"),
	}
}

func (InvoiceApplicationItem) Edges() []ent.Edge {
	return []ent.Edge{
		// 必须用 .Field("application_id") 复用下面显式声明的字段。
		// 若不指定 .Field()，ent 会自行创建外键列并生成同名 Set 方法，
		// 与显式字段冲突，导致 SetApplicationID 等方法重复声明、生成代码无法编译。
		edge.From("application", InvoiceApplication.Type).
			Ref("items").
			Unique().
			Required().
			Field("application_id"),
	}
}

func (InvoiceApplicationItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_id"),
		index.Fields("released_at"),
		// 一笔充值记录至多被一条未释放的申请占用。
		// 部分唯一索引由手写迁移 248_invoice_applications.sql 落地。
		// 注意：不要再单独声明 index.Fields("order_id")，ent 会把同字段的两个索引
		// 拆成「非唯一带 Where」与「唯一不带 Where」两条，反而丢失部分唯一语义。
		index.Fields("order_id").
			Annotations(entsql.IndexWhere("released_at IS NULL")).
			Unique(),
	}
}
