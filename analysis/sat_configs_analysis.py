import pandas as pd
import matplotlib.pyplot as plt

csv_file = 'results/uf100-430_pa.csv'

try:
    df = pd.read_csv(csv_file)
except FileNotFoundError:
    print(f"Error: File {csv_file} not found")
    exit(1)
except pd.errors.EmptyDataError:
    print(f"Error: File {csv_file} is empty")
    exit(1)

if 'Config' not in df.columns:
    print(f"Error: Column 'Config' not found in {csv_file}")
    exit(1)

config_counts = df['Config'].value_counts()

print("Config frequencies:")
for config, count in config_counts.items():
    print(f"{config}: {count}")

plt.figure(figsize=(10, 6))
config_counts.plot(kind='bar', edgecolor='black', alpha=0.7, color='skyblue')
plt.title(f'Frequency of Configurations ({csv_file})')
plt.xlabel('Configuration')
plt.ylabel('Frequency')
plt.grid(True, alpha=0.3)
plt.xticks(rotation=45, ha='right')

output_png = f"{csv_file.replace('.csv', '_config_histogram.png')}"
plt.savefig(output_png, bbox_inches='tight')
print(f"Histogram saved to {output_png}")
plt.show()