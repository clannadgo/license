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
      <!-- License统计图表 -->
      <div class="license-chart">
        <div id="licenseChart"></div>
      </div>

      <!-- License列表 -->
      <div class="license-table">
        <div class="table-header">
          <h3>License列表</h3>
        </div>
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
        <!-- 分页组件 -->
        <div class="pagination-container">
          <div id="paginationChart" class="pagination-chart"></div>
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
const paginationChartInstance = ref(null)
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

// 初始化图表
const initChart = () => {
  const chartDom = document.getElementById('licenseChart')
  if (chartDom) {
    chartInstance.value = echarts.init(chartDom)
    updateChart()
  }
}

// 更新图表数据
const updateChart = () => {
  if (!chartInstance.value) return
  
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
    updatePaginationChart()
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



// 初始化分页图表
const initPaginationChart = () => {
  const chartDom = document.getElementById('paginationChart')
  if (chartDom) {
    paginationChartInstance.value = echarts.init(chartDom)
    updatePaginationChart()
  }
}

// 更新分页图表
const updatePaginationChart = () => {
  if (!paginationChartInstance.value) return
  
  const totalPages = Math.ceil(total.value / pageSize.value)
  const pageSizes = [10, 20, 50, 100]
  
  // 计算显示的页码范围
  let startPage = Math.max(1, currentPage.value - 2)
  let endPage = Math.min(totalPages, startPage + 4)
  if (endPage - startPage < 4 && startPage > 1) {
    startPage = Math.max(1, endPage - 4)
  }
  
  // 生成页码数据
  const pageItems = []
  for (let i = startPage; i <= endPage; i++) {
    pageItems.push({
      name: i.toString(),
      value: i,
      itemStyle: {
        color: i === currentPage.value ? '#1890ff' : '#f0f0f0'
      }
    })
  }
  
  // 生成每页条数选项
  const sizeItems = pageSizes.map(size => ({
    name: size.toString(),
    value: size,
    itemStyle: {
      color: size === pageSize.value ? '#1890ff' : '#f0f0f0'
    }
  }))
  
  const option = {
    backgroundColor: 'transparent',
    animation: false,
    grid: {
      left: '3%',
      right: '3%',
      top: '10%',
      bottom: '10%',
      containLabel: true
    },
    tooltip: {
      show: false
    },
    graphic: [
      // 总数显示
      {
        type: 'text',
        left: 'left',
        top: 'middle',
        style: {
          text: `共 ${total.value} 条`,
          fill: '#333',
          fontSize: 12
        }
      },
      // 每页条数选择
      {
        type: 'group',
        left: '20%',
        top: 'middle',
        children: sizeItems.map((item, index) => ({
          type: 'rect',
          shape: {
            x: index * 40,
            y: -15,
            width: 30,
            height: 30,
            r: 4
          },
          style: {
            fill: item.itemStyle.color
          },
          onclick: () => {
            if (item.value !== pageSize.value) {
              pageSize.value = item.value
              currentPage.value = 1
              fetchLicenseList(currentPage.value, pageSize.value)
            }
          }
        })).concat(sizeItems.map((item, index) => ({
          type: 'text',
          left: index * 40 + 15,
          top: 0,
          style: {
            text: item.name,
            fill: '#fff',
            fontSize: 12,
            textVerticalAlign: 'middle',
            textAlign: 'center'
          },
          onclick: () => {
            if (item.value !== pageSize.value) {
              pageSize.value = item.value
              currentPage.value = 1
              fetchLicenseList(currentPage.value, pageSize.value)
            }
          }
        })))
      },
      // 上一页按钮
      {
        type: 'rect',
        left: '45%',
        top: 'middle',
        shape: {
          x: -20,
          y: -15,
          width: 40,
          height: 30,
          r: 4
        },
        style: {
          fill: currentPage.value > 1 ? '#1890ff' : '#f0f0f0'
        },
        onclick: () => {
          if (currentPage.value > 1) {
            currentPage.value--
            fetchLicenseList(currentPage.value, pageSize.value)
          }
        }
      },
      {
        type: 'text',
        left: '45%',
        top: 'middle',
        style: {
          text: '上一页',
          fill: currentPage.value > 1 ? '#fff' : '#666',
          fontSize: 12,
          textVerticalAlign: 'middle',
          textAlign: 'center'
        },
        onclick: () => {
          if (currentPage.value > 1) {
            currentPage.value--
            fetchLicenseList(currentPage.value, pageSize.value)
          }
        }
      },
      // 页码按钮
      {
        type: 'group',
        left: '55%',
        top: 'middle',
        children: pageItems.map((item, index) => ({
          type: 'rect',
          shape: {
            x: index * 40,
            y: -15,
            width: 30,
            height: 30,
            r: 4
          },
          style: {
            fill: item.itemStyle.color
          },
          onclick: () => {
            if (item.value !== currentPage.value) {
              currentPage.value = item.value
              fetchLicenseList(currentPage.value, pageSize.value)
            }
          }
        })).concat(pageItems.map((item, index) => ({
          type: 'text',
          left: index * 40 + 15,
          top: 0,
          style: {
            text: item.name,
            fill: item.value === currentPage.value ? '#fff' : '#333',
            fontSize: 12,
            textVerticalAlign: 'middle',
            textAlign: 'center'
          },
          onclick: () => {
            if (item.value !== currentPage.value) {
              currentPage.value = item.value
              fetchLicenseList(currentPage.value, pageSize.value)
            }
          }
        })))
      },
      // 下一页按钮
      {
        type: 'rect',
        left: '75%',
        top: 'middle',
        shape: {
          x: -20,
          y: -15,
          width: 40,
          height: 30,
          r: 4
        },
        style: {
          fill: currentPage.value < totalPages ? '#1890ff' : '#f0f0f0'
        },
        onclick: () => {
          if (currentPage.value < totalPages) {
            currentPage.value++
            fetchLicenseList(currentPage.value, pageSize.value)
          }
        }
      },
      {
        type: 'text',
        left: '75%',
        top: 'middle',
        style: {
          text: '下一页',
          fill: currentPage.value < totalPages ? '#fff' : '#666',
          fontSize: 12,
          textVerticalAlign: 'middle',
          textAlign: 'center'
        },
        onclick: () => {
          if (currentPage.value < totalPages) {
            currentPage.value++
            fetchLicenseList(currentPage.value, pageSize.value)
          }
        }
      },
      // 跳转输入框 (模拟)
      {
        type: 'group',
        left: '85%',
        top: 'middle',
        children: [
          {
            type: 'text',
            left: 0,
            top: 0,
            style: {
              text: `跳至 ${currentPage.value}`,
              fill: '#333',
              fontSize: 12,
              textVerticalAlign: 'middle',
              textAlign: 'center'
            }
          },
          {
            type: 'rect',
            shape: {
              x: 40,
              y: -15,
              width: 50,
              height: 30,
              r: 4
            },
            style: {
              fill: '#1890ff'
            },
            onclick: () => {
              // 简单实现，实际应该有输入框
              const pageNum = prompt('请输入页码:', currentPage.value)
              if (pageNum && !isNaN(pageNum)) {
                const num = parseInt(pageNum)
                if (num >= 1 && num <= totalPages) {
                  currentPage.value = num
                  fetchLicenseList(currentPage.value, pageSize.value)
                }
              }
            }
          },
          {
            type: 'text',
            left: 65,
            top: 0,
            style: {
              text: 'GO',
              fill: '#fff',
              fontSize: 12,
              textVerticalAlign: 'middle',
              textAlign: 'center'
            },
            onclick: () => {
              // 简单实现，实际应该有输入框
              const pageNum = prompt('请输入页码:', currentPage.value)
              if (pageNum && !isNaN(pageNum)) {
                const num = parseInt(pageNum)
                if (num >= 1 && num <= totalPages) {
                  currentPage.value = num
                  fetchLicenseList(currentPage.value, pageSize.value)
                }
              }
            }
          }
        ]
      }
    ]
  }
  
  paginationChartInstance.value.setOption(option, { replace: true })
}

// 分页变化处理
const handlePageChange = (page, size) => {
  currentPage.value = page
  pageSize.value = size
  fetchLicenseList(page, size)
}

// 组件挂载时初始化
onMounted(() => {
  fetchLicenseList(currentPage.value, pageSize.value)
  initChart()
  initPaginationChart()
  
  // 窗口大小变化时重新渲染图表
  window.addEventListener('resize', () => {
    if (chartInstance.value) {
      chartInstance.value.resize()
    }
    if (paginationChartInstance.value) {
      paginationChartInstance.value.resize()
    }
  })
})
</script>

<style scoped>
.license-container {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
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

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
  min-height: 80px;
}

.pagination-chart {
  width: 100%;
  height: 80px;
}

.license-table:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

#licenseChart {
  width: 100%;
  height: 250px; /* 减小图表高度 */
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