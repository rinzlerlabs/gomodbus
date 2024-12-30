package transport

import (
	"context"
	"fmt"
)

type TransactionManager interface {
	SendRequest(ctx context.Context, frame ApplicationDataUnit) (ApplicationDataUnit, error)
	WriteResponse(adu ApplicationDataUnit, pdu *ProtocolDataUnit) error
	Close() error
}

type transactionManager struct {
	transport    Transport
	frameBuilder FrameBuilder
}

func NewTransactionManager(transport Transport, frameBuilder FrameBuilder) TransactionManager {
	return &transactionManager{
		transport:    transport,
		frameBuilder: frameBuilder,
	}
}

func (tm *transactionManager) SendRequest(ctx context.Context, adu ApplicationDataUnit) (ApplicationDataUnit, error) {
	err := tm.transport.WriteFrame(adu)
	if err != nil {
		return nil, fmt.Errorf("failed to send request frame: %w", err)
	}
	response, err := tm.transport.ReadResponse(ctx, adu)
	if err != nil {
		return nil, fmt.Errorf("failed to read response frame: %w", err)
	}
	return response, nil
}

func (tm *transactionManager) WriteResponse(adu ApplicationDataUnit, pdu *ProtocolDataUnit) error {
	responseFrame := tm.frameBuilder.BuildResponseFrame(adu, pdu)
	err := tm.transport.WriteFrame(responseFrame)
	if err != nil {
		return fmt.Errorf("failed to write response frame: %w", err)
	}
	return nil
}

func (tm *transactionManager) Close() error {
	return tm.transport.Close()
}
