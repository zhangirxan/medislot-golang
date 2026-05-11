CREATE UNIQUE INDEX IF NOT EXISTS idx_appt_slot_active_unique
ON appointments(slot_id)
WHERE status IN ('pending', 'confirmed');
