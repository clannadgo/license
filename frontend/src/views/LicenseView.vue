<template>
  <div class="license-container">
    <div class="header">
      <h1>License管理系统</h1>
    </div>
    
    <div class="license-actions">
      <el-button type="primary" @click="showAddDialog = true">新增License</el-button>
      <el-button @click="refreshData">刷新数据</el-button>
    </div>

    <div class="license-content">
      <div class="content-row">
        <!-- License统计图表 - 小模块 -->
        <div class="license-chart-small">
          <div id="licenseChart"></div>
        </div>
        
        <!-- License列表 -->
        <div class="license-table">
          <div class="table-header">
            <h3>License列表</h3>
          </div>
          <div class="table-container">
            <el-table :data="licenseList" style="width: 100%">
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column prop="customer" label="客户名称" width="150" />
              <el-table-column prop="fingerprint" label="硬件指纹" width="180" />
              <el-table-column prop="activated_at" label="激活时间" width="180">
                <template #default="scope">
                  {{ formatDate(scope.row.activated_at) }}
                </template>
              </el-table-column>
              <el-table-column prop="expires_at" label="过期时间" width="180">
                <template #default="scope">
                  {{ formatDate(scope.row.expires_at) }}
                </template>
              </el-table-column>
              <el-table-column prop="is_active" label="状态" width="100">
                <template #default="scope">
                  <el-tag :type="scope.row.is_active ? 'success' : 'danger'">
                    {{ scope.row.is_active ? '激活' : '停用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="150">
                <template #default="scope">
                  <el-button size="small" @click="deactivateLicense(scope.row.id)" v-if="scope.row.is_active">
                    停用
                  </el-button>
                  <el-button size="small" type="danger" @click="deleteLicense(scope.row.id)">
                    删除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <!-- 分页组件 -->
          <div class="pagination-container">
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="total"
              layout="total, sizes, prev, pager, next, jumper"
              prev-text="上一页"
              next-text="下一页"
              :pager-count="7"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 新增License对话框 -->
    <el-dialog v-model="showAddDialog" title="新增License" width="500px">
      <el-form :model="newLicense" label-width="120px">
        <el-form-item label="客户名称">
          <el-input v-model="newLicense.customer" placeholder="请输入客户名称" />
        </el-form-item>
        <el-form-item label="硬件指纹">
          <el-input v-model="newLicense.fingerprint" placeholder="请输入硬件指纹" />
        </el-form-item>
        <el-form-item label="有效期(天)">
          <el-input-number v-model="newLicense.validityDays" :min="1" :max="3650" />
        </el-form-item>
        <el-form-item label="License内容">
          <el-input v-model="newLicense.licenseContent" type="textarea" rows="4" placeholder="请输入License内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddDialog = false">取消</el-button>
          <el-button type="primary" @click="addLicense">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

// 数据定义
const licenseList = ref([])
const showAddDialog = ref(false)
const chartInstance = ref(null)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

const newLicense = ref({
  customer: '',
  fingerprint: '',
  validityDays: 365,
  licenseContent: ''
})

// 日期格式化函数
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  if (isNaN(date.getTime())) return dateString
  
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// API基础URL
const API_BASE_URL = 'http://localhost:8080/api'

// 更新图表数据
const updateChart = () => {
  // 销毁旧实例并重新创建，避免type变化的冲突
  if (chartInstance.value) {
    chartInstance.value.dispose()
    chartInstance.value = null
  }
  
  const chartDom = document.getElementById('licenseChart')
  if (!chartDom) return
  
  chartInstance.value = echarts.init(chartDom)
  
  // 计算统计数据
  const totalCount = licenseList.value.length
  const activeCount = licenseList.value.filter(item => item.is_active).length
  const expiredCount = totalCount - activeCount
  
  const option = {
    title: {
      text: 'License统计',
      left: 'center'
    },
    tooltip: {
      trigger: 'item'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: 'License状态',
        type: 'pie',
        radius: '50%',
        data: [
          { value: activeCount, name: '激活中' },
          { value: expiredCount, name: '已过期/停用' }
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  }
  
  chartInstance.value.setOption(option)
}

// 获取License列表
const fetchLicenseList = async (page = 1, size = 10) => {
  try {
    const response = await axios.get(`${API_BASE_URL}/license/activations`, {
      params: {
        page,
        size
      }
    })
    
    // 处理后端返回的数据
    // 假设后端可能支持分页或不分页的两种情况
    if (response.data.activations) {
      licenseList.value = response.data.activations || []
      // 如果后端支持分页，会返回总数
      if (response.data.total !== undefined) {
        total.value = response.data.total
      } else {
        total.value = licenseList.value.length
      }
    } else {
      // 兼容不分页的情况
      licenseList.value = response.data || []
      total.value = licenseList.value.length
    }
    
    updateChart()
  } catch (error) {
    console.error('获取License列表失败:', error)
    ElMessage.error('获取License列表失败')
  }
}

// 添加License
const addLicense = async () => {
  try {
    // 首先激活License
    const activateResponse = await axios.post(`${API_BASE_URL}/license/activate`, {
      license: newLicense.value.licenseContent
    })
    
    if (activateResponse.data.success) {
      ElMessage.success('License添加成功')
      showAddDialog.value = false
      // 重置表单
      newLicense.value = {
        customer: '',
        fingerprint: '',
        validityDays: 365,
        licenseContent: ''
      }
      // 刷新列表
      fetchLicenseList()
    } else {
      ElMessage.error(activateResponse.data.message || 'License添加失败')
    }
  } catch (error) {
    console.error('添加License失败:', error)
    ElMessage.error('添加License失败')
  }
}

// 停用License
const deactivateLicense = async (id) => {
  try {
    await ElMessageBox.confirm('确定要停用这个License吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 这里需要根据后端API实现停用功能
    // 目前后端可能没有提供停用特定License的API
    ElMessage.info('停用功能需要后端API支持')
  } catch (error) {
    // 用户取消操作
  }
}

// 删除License
const deleteLicense = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个License吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 这里需要根据后端API实现删除功能
    // 目前后端可能没有提供删除特定License的API
    ElMessage.info('删除功能需要后端API支持')
  } catch (error) {
    // 用户取消操作
  }
}

// 刷新数据
const refreshData = () => {
  fetchLicenseList(currentPage.value, pageSize.value)
  ElMessage.success('数据已刷新')
}



// 每页条数变化处理
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  fetchLicenseList(currentPage.value, pageSize.value)
}

// 当前页变化处理
const handleCurrentChange = (page) => {
  currentPage.value = page
  fetchLicenseList(currentPage.value, pageSize.value)
}

// 组件挂载时初始化
onMounted(() => {
  fetchLicenseList(currentPage.value, pageSize.value)
  
  // 窗口大小变化时重新渲染图表
  window.addEventListener('resize', () => {
    // 使用setTimeout避免在main process期间调用resize
    setTimeout(() => {
      if (chartInstance.value && !chartInstance.value.isDisposed()) {
        chartInstance.value.resize()
      }
    }, 0)
  })
})
</script>

<style scoped>
.license-container {
  padding: 20px;
  width: 100%;
  margin: 0;
  min-height: 100vh;
  background-color: #f5f7fa;
}

.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 30px 0;
  text-align: center;
  border-radius: 8px;
  margin-bottom: 30px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
}

h1 {
  margin: 0;
  font-size: 2.2rem;
  font-weight: 600;
  letter-spacing: -0.5px;
}

.license-actions {
  margin-bottom: 20px;
  display: flex;
  gap: 15px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
}

.license-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.content-row {
  display: flex;
  gap: 24px;
  align-items: stretch; /* 改为stretch使子元素高度一致 */
}

.license-chart-small {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  padding: 20px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  width: 350px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
}

.license-chart-small:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

.license-chart {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  padding: 24px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.license-chart:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

.license-table {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  padding: 24px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.table-container {
  flex: 1; /* 表格容器填充剩余空间 */
  overflow: auto; /* 添加滚动条以防内容过多 */
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  flex-shrink: 0; /* 防止分页组件被压缩 */
}

.table-header {
  margin-bottom: 20px;
  border-bottom: 2px solid #f0f0f0;
  padding-bottom: 10px;
}

.table-header h3 {
  margin: 0;
  color: #333;
  font-size: 1.2rem;
  font-weight: 600;
}

.license-table:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

#licenseChart {
  width: 100%;
  flex: 1; /* 使图表填充剩余空间 */
  min-height: 300px; /* 设置最小高度 */
}

/* 响应式设计 */
@media (max-width: 768px) {
  .license-container {
    padding: 15px;
  }
  
  .header {
    padding: 20px 0;
  }
  
  h1 {
    font-size: 1.8rem;
  }
  
  .license-actions {
    flex-direction: column;
    gap: 10px;
    padding: 15px;
  }
  
  .license-chart,
  .license-table {
    padding: 15px;
  }
  
  #licenseChart {
    height: 200px;
  }
  
  .pagination-container {
    justify-content: center;
  }
}
</style>