package format

import (
	. "gopkg.in/check.v1"
)

func (s *FormatSuite) TestAutomatic_CorrectDetectionCisco(c *C) {

	find := `<189>571: hostname: Nov  8 13:53:12.226: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1)`
	detected := detect([]byte(find))
	c.Assert(detected, Equals, detectedRFC3164)
}
