-- 开票申请：用户勾选已充值订单提交开票申请，管理员审核并登记发票信息。
--
-- 关联策略：不建指向 users / payment_orders 的外键。
--   - payment_orders 使用硬删除，建外键会阻断订单删除，改变既有行为；
--   - 用户被删除后开票凭证仍需留存以备审计。
-- 因此用户邮箱、用户名与订单号、实付金额、支付时间均以快照形式冗余存储。
CREATE TABLE IF NOT EXISTS invoice_applications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    user_email VARCHAR(255) NOT NULL DEFAULT '',
    user_name VARCHAR(100) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    invoice_type VARCHAR(30) NOT NULL DEFAULT 'electron_normal',
    title_type VARCHAR(20) NOT NULL DEFAULT 'personal',
    title VARCHAR(200) NOT NULL,
    tax_number VARCHAR(64) NOT NULL DEFAULT '',
    registered_address VARCHAR(255) NOT NULL DEFAULT '',
    registered_phone VARCHAR(50) NOT NULL DEFAULT '',
    bank_name VARCHAR(200) NOT NULL DEFAULT '',
    bank_account VARCHAR(100) NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL,
    remark TEXT NOT NULL DEFAULT '',
    total_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    invoice_amount DECIMAL(20,2),
    invoice_number VARCHAR(64) NOT NULL DEFAULT '',
    invoice_code VARCHAR(64) NOT NULL DEFAULT '',
    issued_at TIMESTAMPTZ,
    reviewed_by BIGINT,
    reviewed_at TIMESTAMPTZ,
    rejected_reason TEXT NOT NULL DEFAULT '',
    admin_remark TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_applications_user_id
    ON invoice_applications (user_id);

CREATE INDEX IF NOT EXISTS idx_invoice_applications_user_status
    ON invoice_applications (user_id, status);

CREATE INDEX IF NOT EXISTS idx_invoice_applications_status
    ON invoice_applications (status);

CREATE INDEX IF NOT EXISTS idx_invoice_applications_created_at
    ON invoice_applications (created_at);

-- 申请明细：记录一条申请占用了哪一笔充值订单。
-- released_at IS NULL 表示该充值记录仍被占用；驳回或取消时写入 released_at 释放占用，
-- 明细行保留以供审计。
CREATE TABLE IF NOT EXISTS invoice_application_items (
    id BIGSERIAL PRIMARY KEY,
    application_id BIGINT NOT NULL REFERENCES invoice_applications (id) ON DELETE CASCADE,
    order_id BIGINT NOT NULL,
    order_no VARCHAR(64) NOT NULL DEFAULT '',
    amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    paid_at TIMESTAMPTZ,
    released_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_application_items_application_id
    ON invoice_application_items (application_id);

CREATE INDEX IF NOT EXISTS idx_invoice_application_items_order_id
    ON invoice_application_items (order_id);

CREATE INDEX IF NOT EXISTS idx_invoice_application_items_released_at
    ON invoice_application_items (released_at);

-- 一笔充值记录至多被一条未释放的开票申请占用。
-- 该约束落在数据库层而非仅靠应用层判断，并发提交也无法让同一笔充值被两笔申请同时占用。
CREATE UNIQUE INDEX IF NOT EXISTS idx_invoice_application_items_order_unique_active
    ON invoice_application_items (order_id)
    WHERE released_at IS NULL;
