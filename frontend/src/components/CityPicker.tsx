interface Props {
  label: string;
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
}

export default function CityPicker({ label, value, onChange, onSubmit }: Props) {
  return (
    <form
      className="city-picker"
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit();
      }}
    >
      <label>
        {label}
        <input
          value={value}
          onChange={(event) => onChange(event.target.value)}
          placeholder="e.g. Tokyo"
        />
      </label>
      <button type="submit">Look up</button>
    </form>
  );
}
