
CREATE TABLE IF NOT EXISTS user_coupon_usage (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    coupon_code VARCHAR(50) NOT NULL REFERENCES coupons(code) ON DELETE CASCADE,
    order_id   UUID REFERENCES orders(id) ON DELETE SET NULL,
    used_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, coupon_code)
);

CREATE INDEX IF NOT EXISTS idx_user_coupon_usage_user ON user_coupon_usage(user_id);
CREATE INDEX IF NOT EXISTS idx_user_coupon_usage_code ON user_coupon_usage(coupon_code);
