export default function Loader({ text = 'Loading...' }) {
  return (
    <div className="min-h-screen grid place-items-center">
      <p className="text-slate-600">{text}</p>
    </div>
  );
}
