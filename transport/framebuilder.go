package transport

type FrameBuilder interface {
	BuildResponseFrame(header Header, response *ProtocolDataUnit) ApplicationDataUnit
}
