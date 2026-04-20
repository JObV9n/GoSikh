import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import Input from './Input';

describe('Input', () => {
  it('generates a fallback id and associates label when id is not provided', () => {
    render(<Input label="Email" name="email" />);

    const input = screen.getByRole('textbox', { name: 'Email' });
    const label = screen.getByText('Email');

    expect(input.id).toBeTruthy();
    expect(label).toHaveAttribute('for', input.id);
  });

  it('uses provided id when present', () => {
    render(<Input id="custom-email" label="Email" name="email" />);

    const input = screen.getByRole('textbox', { name: 'Email' });
    const label = screen.getByText('Email');

    expect(input).toHaveAttribute('id', 'custom-email');
    expect(label).toHaveAttribute('for', 'custom-email');
  });
});