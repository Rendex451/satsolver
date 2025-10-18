import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import os
import glob

results_dir = 'results'
csv_files = glob.glob(os.path.join(results_dir, '*.csv'))

for csv_file in csv_files:
    try:
        df = pd.read_csv(csv_file)
    except FileNotFoundError:
        print(f"Error: File {csv_file} not found")
        exit(1)
    except pd.errors.EmptyDataError:
        print(f"Error: File {csv_file} is empty")
        exit(1)

    times = df['Elapsed_time_ms'].astype(float)

    median_time = np.median(times)
    mean_time = np.mean(times)
    min_time = np.min(times)
    max_time = np.max(times)

    print(f"Dataset: {csv_file}")
    print(f"Median time: {median_time:.4f} ms")
    print(f"Mean time: {mean_time:.4f} ms")
    print(f"Min time: {min_time:.4f} ms")
    print(f"Max time: {max_time:.4f} ms")
    print()

    plt.figure(figsize=(10, 6))
    plt.hist(times, bins=30, edgecolor='black', alpha=0.7)
    plt.title('Histogram of Elapsed Times')
    plt.xlabel('Elapsed Time (ms)')
    plt.ylabel('Frequency')
    plt.grid(True, alpha=0.3)

    # Добавляем линии для медианы и среднего
    plt.axvline(median_time, color='red', linestyle='--', label=f'Median: {median_time:.2f} ms')
    plt.axvline(mean_time, color='green', linestyle='--', label=f'Mean: {mean_time:.2f} ms')
    plt.legend()

    output_png = f"{os.path.splitext(csv_file)[0]}_histogram.png"
    plt.savefig(output_png)
    # print(f"Histogram saved to {output_png}")
    plt.close()