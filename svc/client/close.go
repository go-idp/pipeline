package client

func (c *client) Close() error {
	if c.core == nil {
		return nil
	}

	return c.core.Close()
}
