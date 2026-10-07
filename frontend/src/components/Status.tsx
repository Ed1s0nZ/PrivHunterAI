export const labels: Record<string, string> = {
  true: "疑似越权",
  false: "未发现越权",
  unknown: "待分析",
  error: "检测失败",
  pending: "待复核",
  confirmed: "已确认",
  false_positive: "误报",
  resolved: "已解决",
};
export function Badge({ value }: { value: string }) {
  return (
    <span className={"badge badge-" + value}>{labels[value] || value}</span>
  );
}
export function QueryState({
  loading,
  error,
  empty = false,
  retry,
}: {
  loading: boolean;
  error: string;
  empty?: boolean;
  retry: () => void;
}) {
  if (loading) return <div className="empty">正在加载…</div>;
  if (error)
    return (
      <div className="empty">
        <p className="error" role="alert">
          {error}
        </p>
        <button onClick={retry}>重新加载</button>
      </div>
    );
  if (empty)
    return (
      <div className="empty">
        <b>暂无记录</b>
        <p>符合条件的数据将在这里显示。</p>
      </div>
    );
  return null;
}
